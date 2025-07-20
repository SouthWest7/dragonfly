/*
 *     Copyright 2025 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package job

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"io"

	commonv2 "d7y.io/api/v2/pkg/apis/common/v2"
	dfdaemonv2 "d7y.io/api/v2/pkg/apis/dfdaemon/v2"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	dfdaemonclient "d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/scheduler/resource/standard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Distribute manages file distribution with event-driven scheduling
type Distribute struct {
	// Resource is the resource manager for the scheduler.
	resource standard.Resource

	// File is the file to be distributed.
	file *standard.File

	// Events is the set of events.
	events *sync.Map

	// DownloadQueue is the queue of host IDs to be downloaded.
	downloadQueue []string

	// UploadQueue is the map of block sources.
	uploadQueue map[int32][]string

	// RateLimit is the rate limit for the scheduler.
	rateLimit uint64

	// BandwidthUseRatio is the ratio of bandwidth to be used.
	bandwidthUseRatio float64

	// DownloadQueueMu is the mutex for download queue operations.
	downloadQueueMu sync.RWMutex

	// UploadQueueMu is the mutex for upload queue operations.
	uploadQueueMu sync.RWMutex
}

// NewDistribute creates a new event-driven scheduler
func NewDistribute(resource standard.Resource, file *standard.File) *Distribute {
	return &Distribute{
		resource:          resource,
		file:              file,
		events:            &sync.Map{},
		downloadQueue:     make([]string, 0),
		uploadQueue:       make(map[int32][]string),
		rateLimit:         10 * 1024 * 1024,
		bandwidthUseRatio: 0.8,
	}
}

// Serve starts the event-driven file distribution
func (es *Distribute) Serve(ctx context.Context, fileID string, taskID string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string) error {
	if es.file == nil || es.file.ID != fileID {
		return fmt.Errorf("file %s not found in scheduler", fileID)
	}

	logger.Infof("Starting event-driven file distribution for %s with %d blocks", fileID, es.file.TotalBlocks)

	// Initialize scheduler state
	es.createUploadQueue()
	es.createDownloadQueue()

	// Main event-driven scheduling loop
	return es.run(ctx, taskID, headers, pieceLength, contentForCalculatingTaskId)
}

// run executes the main event-driven scheduling algorithm
func (es *Distribute) run(ctx context.Context, taskID string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string) error {
	for {
		// Check if distribution is finished
		if es.isDistributionFinished() {
			logger.Infof("File distribution completed")
			return nil
		}

		// Check if download queue is empty and no events are running
		es.downloadQueueMu.RLock()
		downloadQueueEmpty := len(es.downloadQueue) == 0
		es.downloadQueueMu.RUnlock()

		if downloadQueueEmpty {
			logger.Infof("No more downloads to schedule and no events running")
			return nil
		}

		es.download(ctx, taskID, headers, pieceLength, contentForCalculatingTaskId)
	}
}

// download schedules the next download
func (es *Distribute) download(ctx context.Context, taskID string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string) bool {

	if len(es.downloadQueue) == 0 {
		return false
	}

	// Take the first element from the queue (FIFO)
	hostID := es.downloadQueue[0]

	if es.canScheduleDownload(hostID) {
		blockNumber := es.selectBlockForPeer(hostID)
		if blockNumber != -1 {
			eventID := standard.EventID(hostID, es.file.ID, blockNumber)
			event := es.createEvent(eventID, blockNumber, taskID, hostID)
			if event != nil {
				es.StoreEvent(event)
				// Remove the first element from the queue
				es.downloadQueueMu.Lock()
				es.downloadQueue = es.downloadQueue[1:]
				es.downloadQueueMu.Unlock()

				go es.downloadBlock(ctx, event, taskID, headers, pieceLength, contentForCalculatingTaskId)

				logger.Debugf("Scheduled download: block %d to peer %s", blockNumber, hostID)
				return true
			}
		}
	}

	return false
}

// selectBlockForPeer selects the best block for a peer
func (es *Distribute) selectBlockForPeer(hostID string) int32 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return -1
	}

	file, found := host.LoadFile(es.file.ID)
	if !found {
		return -1
	}

	missingBlocks := file.GetMissingBlocks()
	if len(missingBlocks) == 0 {
		return -1
	}

	var bestBlock int32 = -1
	minSources := int(^uint(0) >> 1)

	es.uploadQueueMu.RLock()
	defer es.uploadQueueMu.RUnlock()

	for _, blockNumber := range missingBlocks {
		if sources, exists := es.uploadQueue[blockNumber]; exists {
			if file.IsBlockAssigned(blockNumber, hostID) {
				continue
			}

			if len(sources) < minSources && len(sources) > 0 {
				minSources = len(sources)
				bestBlock = blockNumber
			}
		}
	}

	return bestBlock
}

// downloadBlock executes the download
func (es *Distribute) downloadBlock(ctx context.Context, event *standard.Event, taskID string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string) {
	host, exists := es.resource.HostManager().Load(event.DownloadPeer)
	if !exists {
		logger.Errorf("Host %s not found for download", event.DownloadPeer)
		return
	}

	file, found := host.LoadFile(es.file.ID)
	if !found {
		logger.Errorf("File %s not found on host %s", es.file.ID, event.DownloadPeer)
		return
	}

	// Assign block to peer in the file system
	_, err := file.AssignBlockToPeer(event.BlockNumber, event.DownloadPeer)
	if err != nil {
		logger.Warnf("Failed to assign block %d to peer %s: %s", event.BlockNumber, event.DownloadPeer, err.Error())
		return
	}

	// Get block information to create Range
	block, exists := file.GetBlock(event.BlockNumber)
	if !exists {
		logger.Errorf("Block %d not found in file %s", event.BlockNumber, es.file.ID)
		return
	}

	// Create commonv2.Range directly instead of HTTP header
	blockRange := &commonv2.Range{
		Start:  uint64(block.Offset),
		Length: uint64(block.Length),
	}

	// Create request headers without Range header
	requestHeaders := make(map[string]string)
	for k, v := range headers {
		requestHeaders[k] = v
	}

	// Connect and download
	addr := fmt.Sprintf("%s:%d", host.IP, host.Port)
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	client, err := dfdaemonclient.GetV2ByAddr(ctx, addr, dialOptions...)
	if err != nil {
		logger.Errorf("Failed to connect to %s: %s", addr, err.Error())
		return
	}

	downloadReq := &dfdaemonv2.DownloadTaskRequest{
		Download: &commonv2.Download{
			Url:           taskID,
			PieceLength:   pieceLength,
			Type:          commonv2.TaskType_STANDARD,
			Priority:      commonv2.Priority_LEVEL0,
			RequestHeader: requestHeaders,
			Range:         blockRange,
		},
	}

	stream, err := client.DownloadTask(ctx, taskID, downloadReq)
	if err != nil {
		logger.Errorf("Failed to start download: %s", err.Error())
		return
	}

	// Wait for completion
	for {
		_, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logger.Infof("Download completed for block %d, peer %s", event.BlockNumber, event.DownloadPeer)
				break
			}
			logger.Errorf("Download failed for block %d, peer %s: %s", event.BlockNumber, event.DownloadPeer, err.Error())
			return
		}
	}

	es.releaseRate(event.DownloadPeer, event.AllocatedDownloadRate, false)

	// Update queues with proper locking
	es.createDownloadQueue()
	es.createUploadQueue()
}

// createUploadQueue creates the upload queue for the scheduler.
func (es *Distribute) createUploadQueue() {
	es.uploadQueueMu.Lock()
	defer es.uploadQueueMu.Unlock()

	es.uploadQueue = make(map[int32][]string)

	// For each block, find hosts that have it
	for i := int32(0); i < int32(es.file.TotalBlocks); i++ {
		var sources []string
		es.resource.HostManager().Range(func(key, value any) bool {
			if hostID, ok := key.(string); ok {
				if host, ok := value.(*standard.Host); ok {
					if file, found := host.LoadFile(es.file.ID); found {
						if file.FinishedBlocks.Test(uint(i)) {
							sources = append(sources, hostID)
						}
					}
				}
			}
			return true
		})

		// Sort sources by upload capacity (descending)
		sort.Slice(sources, func(i, j int) bool {
			rateI := es.getHostIdleUploadRate(sources[i])
			rateJ := es.getHostIdleUploadRate(sources[j])
			return rateI > rateJ
		})

		if len(sources) > 0 {
			es.uploadQueue[i] = sources
		}
	}
}

// createDownloadQueue creates the download queue for the scheduler.
func (es *Distribute) createDownloadQueue() {
	es.downloadQueueMu.Lock()
	defer es.downloadQueueMu.Unlock()

	es.downloadQueue = make([]string, 0)

	es.resource.HostManager().Range(func(key, value any) bool {
		if hostID, ok := key.(string); ok {
			if host, ok := value.(*standard.Host); ok {
				if file, found := host.LoadFile(es.file.ID); found {
					missing := file.GetMissingBlockCount()
					// Check if host has missing blocks and sufficient download bandwidth
					if missing > 0 && es.getHostIdleDownloadRate(hostID) > es.rateLimit {
						es.downloadQueue = append(es.downloadQueue, hostID)
					}
				}
			}
		}
		return true
	})

	// Sort download queue by missing block count (descending)
	sort.Slice(es.downloadQueue, func(i, j int) bool {
		hostI, existsI := es.resource.HostManager().Load(es.downloadQueue[i])
		hostJ, existsJ := es.resource.HostManager().Load(es.downloadQueue[j])

		if !existsI || !existsJ {
			return false
		}

		fileI, foundI := hostI.LoadFile(es.file.ID)
		fileJ, foundJ := hostJ.LoadFile(es.file.ID)

		if !foundI || !foundJ {
			return false
		}

		missingI := fileI.GetMissingBlockCount()
		missingJ := fileJ.GetMissingBlockCount()

		return missingI > missingJ
	})

	logger.Infof("Initialized download queue with %d peers", len(es.downloadQueue))
}

// canScheduleDownload checks if a host can be scheduled for download.
func (es *Distribute) canScheduleDownload(hostID string) bool {
	return es.getHostIdleDownloadRate(hostID) > es.rateLimit
}

// isDistributionFinished checks if the distribution is finished.
func (es *Distribute) isDistributionFinished() bool {
	finished := true
	es.resource.HostManager().Range(func(key, value any) bool {
		if host, ok := value.(*standard.Host); ok {
			if file, found := host.LoadFile(es.file.ID); found {
				if !file.IsFinished() {
					finished = false
					return false
				}
			}
		}
		return true
	})
	return finished
}

// getHostIdleUploadRate gets the idle upload rate for a host.
func (es *Distribute) getHostIdleUploadRate(hostID string) uint64 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return 0
	}
	idle := host.Network.UploadRateLimit - host.Network.AllocatedUploadRate
	if idle < 0 {
		return 0
	}
	return idle
}

// getHostIdleDownloadRate gets the idle download rate for a host.
func (es *Distribute) getHostIdleDownloadRate(hostID string) uint64 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return 0
	}
	idle := host.Network.DownloadRateLimit - host.Network.AllocatedDownloadRate
	if idle < 0 {
		return 0
	}
	return idle
}

// allocateRate allocates the rate for a host.
func (es *Distribute) allocateRate(hostID string, rate uint64, isUpload bool) {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return
	}

	host.AllocateRate(rate, isUpload)
}

// releaseRate releases the rate for a host.
func (es *Distribute) releaseRate(hostID string, rate uint64, isUpload bool) {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return
	}

	host.ReleaseRate(rate, isUpload)
}

// createEvent creates a new event
func (es *Distribute) createEvent(id string, blockNumber int32, taskID string, hostID string) *standard.Event {
	es.uploadQueueMu.RLock()
	defer es.uploadQueueMu.RUnlock()

	parents, exists := es.uploadQueue[blockNumber]
	if !exists || len(parents) == 0 {
		return nil
	}

	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return nil
	}
	file, found := host.LoadFile(es.file.ID)
	if !found {
		return nil
	}

	block, exists := file.GetBlock(blockNumber)
	if !exists {
		return nil
	}

	parentID := parents[0]
	parent, exists := es.resource.HostManager().Load(parentID)
	if !exists {
		return nil
	}

	parentFile, found := parent.LoadFile(es.file.ID)
	if !found {
		return nil
	}

	parentBlock, exists := parentFile.GetBlock(blockNumber)
	if !exists {
		return nil
	}
	block.ParentID = parentBlock.PeerID

	idleUploadRate := es.getHostIdleUploadRate(parentID)
	idleDownloadRate := es.getHostIdleDownloadRate(hostID)

	var allocatedRate uint64

	if idleUploadRate > idleDownloadRate {
		allocatedRate = idleDownloadRate
	} else {
		allocatedRate = idleUploadRate
	}

	es.allocateRate(parentID, allocatedRate, true)
	es.allocateRate(hostID, allocatedRate, false)

	es.createUploadQueue()

	return standard.NewEvent(id, blockNumber, taskID, hostID, parentID, allocatedRate, allocatedRate)
}

func (es *Distribute) LoadEvent(id string) *standard.Event {
	event, ok := es.events.Load(id)
	if !ok {
		return nil
	}

	return event.(*standard.Event)
}

func (es *Distribute) StoreEvent(event *standard.Event) {
	es.events.Store(event.ID, event)
}
