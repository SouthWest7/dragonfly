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
	"time"

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

	// ScheduleInterval is the schedule interval for the scheduler.
	scheduleInterval time.Duration

	// BandwidthUseRatio is the ratio of bandwidth to be used.
	bandwidthUseRatio float64

	// DownloadQueueMu is the mutex for download queue operations.
	downloadQueueMu sync.RWMutex

	// UploadQueueMu is the mutex for upload queue operations.
	uploadQueueMu sync.RWMutex
}

// NewDistribute creates a new event-driven scheduler
func NewDistribute(resource standard.Resource, file *standard.File, rateLimit uint64, scheduleInterval uint64) *Distribute {
	return &Distribute{
		resource:          resource,
		file:              file,
		events:            &sync.Map{},
		downloadQueue:     make([]string, 0),
		uploadQueue:       make(map[int32][]string),
		rateLimit:         rateLimit,
		scheduleInterval:  time.Duration(scheduleInterval) * time.Millisecond,
		bandwidthUseRatio: 0.8,
	}
}

// Run starts the event-driven file distribution
func (es *Distribute) Run(ctx context.Context, fileID string, taskID string, url string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string) error {
	if es.file == nil || es.file.ID != fileID {
		return fmt.Errorf("file %s not found in scheduler", fileID)
	}

	count := 0
	// Main event-driven scheduling loop
	for {
		// Check if distribution is finished
		if es.isDistributionFinished() {
			logger.Infof("[distribute]: file distribution completed, count: %d", count)
			return nil
		}

		es.createUploadQueue(count)
		es.createDownloadQueue(count)

		if len(es.downloadQueue) == 0 {
			logger.Infof("[distribute]: download queue is empty, count: %d", count)
			time.Sleep(es.scheduleInterval)
			count++
			continue
		}

		hostID := es.downloadQueue[0]

		if es.canScheduleDownload(hostID) {
			blockNumber := es.selectBlockForPeer(hostID, count)
			if blockNumber != -1 {
				eventID := standard.EventID(hostID, es.file.ID, blockNumber)
				event := es.createEvent(eventID, blockNumber, taskID, hostID)
				if event != nil {
					es.StoreEvent(event)

					go es.downloadBlock(ctx, event, taskID, url, headers, pieceLength, contentForCalculatingTaskId, count)

					logger.Debugf("[distribute]: scheduled one download block %d to peer %s, count: %d", blockNumber, hostID, count)
				}
			}

			logger.Infof("[distribute]: scheduled one download for peer %s, count: %d", hostID, count)
		}

		time.Sleep(es.scheduleInterval)
		count++
	}
}

// downloadBlock executes the download
func (es *Distribute) downloadBlock(ctx context.Context, event *standard.Event, taskID string, url string, headers map[string]string, pieceLength *uint64, contentForCalculatingTaskId *string, count int) {
	logger.Infof("[distribute]: downloading block %d to peer %s, count: %d", event.BlockNumber, event.DownloadHost, count)

	host, exists := es.resource.HostManager().Load(event.DownloadHost)
	if !exists {
		logger.Errorf("[distribute]: host %s not found for download", event.DownloadHost)
		return
	}

	file, found := host.LoadFile(es.file.ID)
	if !found {
		logger.Errorf("[distribute]: file %s not found on host %s", es.file.ID, event.DownloadHost)
		return
	}

	// Get block information to create Range
	block, exists := file.GetBlock(event.BlockNumber)
	if !exists {
		logger.Errorf("[distribute]: block %d not found in file %s", event.BlockNumber, es.file.ID)
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
		logger.Errorf("[distribute]: failed to connect to %s: %s", addr, err.Error())
		return
	}

	downloadReq := &dfdaemonv2.DownloadTaskRequest{
		Download: &commonv2.Download{
			Url:           url,
			PieceLength:   pieceLength,
			Type:          commonv2.TaskType_STANDARD,
			Priority:      commonv2.Priority_LEVEL0,
			RequestHeader: requestHeaders,
			Range:         blockRange,
		},
	}

	stream, err := client.DownloadTask(ctx, taskID, downloadReq)
	if err != nil {
		logger.Errorf("[distribute]: failed to start download: %s", err.Error())
		return
	}

	// Wait for completion
	for {
		_, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logger.Infof("[distribute]: download completed for block %d, peer %s, count: %d", event.BlockNumber, event.DownloadHost, count)
				break
			}
			block.State = standard.BlockStateFailed
			logger.Errorf("[distribute]: download failed for block %d, peer %s: %s, count: %d", event.BlockNumber, event.DownloadHost, err.Error(), count)
			return
		}
	}

	if err := file.FinishBlock(event.BlockNumber); err != nil {
		logger.Errorf("[distribute]: failed to complete block %d for peer %s: %s, count: %d", event.BlockNumber, event.DownloadHost, err.Error(), count)
	} else {
		block.State = standard.BlockStateCompleted
		logger.Infof("[distribute]: file %s completed block %d for peer %s, count: %d", es.file.ID, event.BlockNumber, event.DownloadHost, count)
	}

	es.releaseRate(event.DownloadHost, event.AllocatedDownloadRate, false)
	es.releaseRate(event.UploadPeers[0], event.AllocatedUploadRate, true)
}

// selectBlockForPeer selects the best block for a peer
func (es *Distribute) selectBlockForPeer(hostID string, count int) int32 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		logger.Errorf("[distribute]: host %s not found for select block", hostID)
		return -1
	}

	file, found := host.LoadFile(es.file.ID)
	if !found {
		logger.Errorf("[distribute]: file %s not found on host %s", es.file.ID, hostID)
		return -1
	}

	missingBlocks := file.GetMissingBlocks()
	if len(missingBlocks) == 0 {
		logger.Infof("[distribute]: no missing blocks for host %s, count: %d", hostID, count)
		return -1
	}

	var bestBlock int32 = -1
	minSources := int(^uint(0) >> 1)

	es.uploadQueueMu.RLock()
	defer es.uploadQueueMu.RUnlock()

	for _, blockNumber := range missingBlocks {
		if sources, exists := es.uploadQueue[blockNumber]; exists {
			if len(sources) < minSources && len(sources) > 0 {
				minSources = len(sources)
				bestBlock = blockNumber
			}
		}
	}

	logger.Infof("[distribute]: select block %d for host %s, count: %d", bestBlock, hostID, count)

	return bestBlock
}

// createUploadQueue creates the upload queue for the scheduler.
func (es *Distribute) createUploadQueue(count int) {
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
						if file.FinishedBlocks.Test(uint(i)) && es.getHostIdleUploadRate(hostID) > es.rateLimit {
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

	logger.Infof("[distribute]: create upload queue %#v, count: %d", es.uploadQueue, count)
}

// createDownloadQueue creates the download queue for the scheduler.
func (es *Distribute) createDownloadQueue(count int) int {
	es.downloadQueueMu.Lock()
	defer es.downloadQueueMu.Unlock()

	es.downloadQueue = make([]string, 0)

	es.resource.HostManager().Range(func(key, value any) bool {
		if hostID, ok := key.(string); ok {
			if host, ok := value.(*standard.Host); ok {
				if file, found := host.LoadFile(es.file.ID); found {
					missing := file.GetMissingBlocks()
					logger.Infof("[distribute]: host %s has %v missing blocks, count: %d", hostID, missing, count)
					// Check if host has missing blocks and sufficient download bandwidth
					if len(missing) > 0 && es.getHostIdleDownloadRate(hostID) > es.rateLimit {
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

		missingI := len(fileI.GetMissingBlocks())
		missingJ := len(fileJ.GetMissingBlocks())

		return missingI > missingJ
	})

	logger.Infof("[distribute]: create download queue with %d peers, count: %d", len(es.downloadQueue), count)

	return len(es.downloadQueue)
}

// createEvent creates a new event
func (es *Distribute) createEvent(id string, blockNumber int32, taskID string, hostID string) *standard.Event {
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

	logger.Infof("[distribute]: allocated rate %d for parent %s and host %s", allocatedRate, parentID, hostID)

	es.allocateRate(parentID, allocatedRate, true)
	es.allocateRate(hostID, allocatedRate, false)

	block.State = standard.BlockStateDownload

	return standard.NewEvent(id, blockNumber, taskID, hostID, parentID, allocatedRate, allocatedRate)
}

// canScheduleDownload checks if a host can be scheduled for download.
func (es *Distribute) canScheduleDownload(hostID string) bool {
	logger.Infof("[distribute]: host %s idle download rate: %d, rate limit: %d", hostID, es.getHostIdleDownloadRate(hostID), es.rateLimit)
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

	occupied := host.Network.UploadRate
	if host.Network.UploadRate < host.Network.AllocatedUploadRate {
		occupied = host.Network.AllocatedUploadRate
	}
	return host.Network.UploadRateLimit - occupied
}

// getHostIdleDownloadRate gets the idle download rate for a host.
func (es *Distribute) getHostIdleDownloadRate(hostID string) uint64 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return 0
	}

	occupied := host.Network.DownloadRate
	if host.Network.DownloadRate < host.Network.AllocatedDownloadRate {
		occupied = host.Network.AllocatedDownloadRate
	}
	return host.Network.DownloadRateLimit - occupied
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
