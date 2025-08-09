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
	"math/rand"
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

	// URL is the URL of the file.
	url string

	// File is the file to be distributed.
	file *standard.File

	// HostNumber is the number of hosts.
	hostNumber int

	// SuperSeed is the super seed for the file as [hostID, taskID, peerID].
	superSeed [3]string

	// SuperSeedEnabled is the super seed mode for the file.
	superSeedEnabled bool

	// DownloadQueue is the queue of host IDs to be downloaded.
	downloadQueue []string

	// UploadQueue is the map of block sources.
	uploadQueue map[int32][]string

	// RateLimit is the rate limit for the peer.
	rateLimit uint64

	// RateThreshold is the rate threshold for the peer.
	rateThreshold uint64

	// Usage is the usage for the rate limit.
	usage float32

	// ScheduleInterval is the schedule interval for the scheduler.
	scheduleInterval time.Duration

	// DownloadQueueMu is the mutex for download queue operations.
	downloadQueueMu sync.RWMutex

	// UploadQueueMu is the mutex for upload queue operations.
	uploadQueueMu sync.RWMutex
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

		workQueue := make([]string, len(es.downloadQueue))
		copy(workQueue, es.downloadQueue)
		for len(workQueue) > 0 {
			hostID := workQueue[0]
			workQueue = workQueue[1:]

			if !es.canScheduleDownload(hostID) {
				continue
			}

			blockNumber := es.selectBlockForPeer(hostID, count)
			logger.Infof("[distribute]: selected block %d for peer %s, count: %d", blockNumber, hostID, count)
			if blockNumber != -1 {
				eventID := standard.EventID(hostID, es.file.ID, blockNumber)
				event := es.createEvent(eventID, blockNumber, taskID, hostID)
				if event != nil {
					go es.downloadBlock(ctx, event, taskID, url, headers, pieceLength, contentForCalculatingTaskId, count)

					es.superSeedMode(blockNumber, count)
					logger.Infof("[distribute]: scheduled one download block %d to peer %s, count: %d", blockNumber, hostID, count)
				}
			}
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
	block, exists := file.GetPartition(event.BlockNumber)
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
			Url:                         url,
			PieceLength:                 pieceLength,
			Type:                        commonv2.TaskType_STANDARD,
			Priority:                    commonv2.Priority_LEVEL0,
			RequestHeader:               requestHeaders,
			Range:                       blockRange,
			ContentForCalculatingTaskId: &es.url,
		},
	}

	logger.Infof("[distribute]: send download request to peer %s, count: %d", event.DownloadHost, count)
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
				es.releaseRate(event.DownloadHost, event.AllocatedDownloadRate, false)
				es.releaseRate(event.UploadPeers[0], event.AllocatedUploadRate, true)
				break
			}
			block.State = standard.PartitionStateFailed
			logger.Errorf("[distribute]: download failed for block %d, peer %s: %s, count: %d", event.BlockNumber, event.DownloadHost, err.Error(), count)
			return
		}
	}

	if err := file.FinishPartition(event.BlockNumber); err != nil {
		logger.Errorf("[distribute]: failed to complete block %d for peer %s: %s, count: %d", event.BlockNumber, event.DownloadHost, err.Error(), count)
	} else {
		block.State = standard.PartitionStateCompleted
		logger.Infof("[distribute]: file %s completed block %d for peer %s, count: %d", es.file.ID, event.BlockNumber, event.DownloadHost, count)
	}
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

	missingBlocks := file.GetMissingPartitions()
	if len(missingBlocks) == 0 {
		logger.Infof("[distribute]: no missing blocks for host %s, count: %d", hostID, count)
		return -1
	}

	var bestBlocks []int32
	minSources := es.hostNumber

	es.uploadQueueMu.RLock()
	defer es.uploadQueueMu.RUnlock()

	for _, blockNumber := range missingBlocks {
		if sources, exists := es.uploadQueue[blockNumber]; exists {
			if len(sources) > 0 {
				if len(sources) < minSources {
					// 找到更少源的块，重置候选列表
					minSources = len(sources)
					bestBlocks = []int32{blockNumber}
				} else if len(sources) == minSources {
					// 找到相同源数量的块，添加到候选列表
					bestBlocks = append(bestBlocks, blockNumber)
				}
			}
		}
	}

	// 如果有候选块，随机选择一个
	if len(bestBlocks) > 0 {
		// 使用时间种子创建随机数生成器确保真随机
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIndex := r.Intn(len(bestBlocks))
		return bestBlocks[randomIndex]
	}

	return -1
}

func (es *Distribute) superSeedMode(partitionNumber int32, count int) {
	// superSeed[0] is hostID, superSeed[1] is taskID, superSeed[2] is peerID
	hostID, taskID, peerID := es.superSeed[0], es.superSeed[1], es.superSeed[2]
	if es.getHostIdleUploadRate(hostID) < es.rateThreshold {
		return
	}

	// Find the host using the hostID and then load the file
	host, found := es.resource.HostManager().Load(hostID)
	if !found {
		return
	}

	file, found := host.LoadFile(es.file.ID)
	if !found {
		return
	}

	if file.FinishedPartitions.Test(uint(partitionNumber)) {
		if file.FinishedPartitions.Count() < es.file.TotalPartitions/3 {
			es.superSeedEnabled = false
			standard.FinishFile(es.resource, hostID, taskID, peerID)
			return
		} else {
			file.FinishedPartitions.Clear(uint(partitionNumber))
			logger.Infof("[distribute]: clear block %d for super seed, count: %d", partitionNumber, count)
		}
	}
}

// createUploadQueue creates the upload queue for the scheduler.
func (es *Distribute) createUploadQueue(count int) {
	es.uploadQueueMu.Lock()
	defer es.uploadQueueMu.Unlock()

	es.uploadQueue = make(map[int32][]string)

	// For each partition, find hosts that have it or downloading it
	for i := int32(0); i < int32(es.file.TotalPartitions); i++ {
		var sources []string
		es.resource.HostManager().Range(func(key, value any) bool {
			if hostID, ok := key.(string); ok {
				if host, ok := value.(*standard.Host); ok {
					if es.getHostIdleUploadRate(hostID) > es.rateThreshold {
						if file, found := host.LoadFile(es.file.ID); found {
							if file.FinishedPartitions.Test(uint(i)) {
								sources = append(sources, hostID)
							} else if partition, exists := file.GetPartition(i); exists {
								peer, exists := es.resource.PeerManager().Load(partition.PeerID)
								if exists && peer.FSM.Is(standard.PeerStateSucceeded) {
									sources = append(sources, hostID)
								}
							}
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
func (es *Distribute) createDownloadQueue(count int) int {
	es.downloadQueueMu.Lock()
	defer es.downloadQueueMu.Unlock()

	es.downloadQueue = make([]string, 0)

	es.resource.HostManager().Range(func(key, value any) bool {
		if hostID, ok := key.(string); ok {
			if hostID == es.superSeed[0] {
				logger.Infof("[distribute]: host %s is super seed, count: %d", hostID, count)
				return true
			}
			if host, ok := value.(*standard.Host); ok {
				if es.getHostIdleDownloadRate(hostID) > es.rateThreshold {
					if file, found := host.LoadFile(es.file.ID); found {
						missing := file.GetMissingPartitions()
						if len(missing) > 0 {
							logger.Infof("[distribute]: host %s has %v missing blocks, idle download rate: %d, rate limit: %d", hostID, missing, es.getHostIdleDownloadRate(hostID), es.rateThreshold)
							es.downloadQueue = append(es.downloadQueue, hostID)
						}
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

		missingI := len(fileI.GetMissingPartitions())
		missingJ := len(fileJ.GetMissingPartitions())

		return missingI > missingJ
	})

	logger.Infof("[distribute]: download queue: %v", es.downloadQueue)

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

	block, exists := file.GetPartition(blockNumber)
	if !exists {
		return nil
	}

	parentID := parents[rand.Intn(len(parents))]
	parent, exists := es.resource.HostManager().Load(parentID)
	if !exists {
		return nil
	}

	parentFile, found := parent.LoadFile(es.file.ID)
	if !found {
		return nil
	}

	parentBlock, exists := parentFile.GetPartition(blockNumber)
	if !exists {
		return nil
	}

	block.ParentID = parentBlock.PeerID

	idleUploadRate := es.getHostIdleUploadRate(parentID)
	logger.Infof("[distribute]: parent %s real upload rate: %d, allocated upload rate: %d", parentID, parent.Network.UploadRate, parent.Network.AllocatedUploadRate)
	idleDownloadRate := es.getHostIdleDownloadRate(hostID)
	logger.Infof("[distribute]: idle upload rate: %d, idle download rate: %d", idleUploadRate, idleDownloadRate)

	allocatedRate := idleUploadRate

	if idleUploadRate > idleDownloadRate {
		allocatedRate = idleDownloadRate
	}
	if allocatedRate > uint64(float64(es.rateLimit)*float64(es.usage)) {
		allocatedRate = uint64(float64(es.rateLimit) * float64(es.usage))
	}

	logger.Infof("[distribute]: allocated rate %d for parent %s and host %s", allocatedRate, parentID, hostID)

	es.allocateRate(parentID, allocatedRate, true)
	es.allocateRate(hostID, allocatedRate, false)

	block.State = standard.PartitionStateDownload

	return standard.NewEvent(id, blockNumber, taskID, hostID, parentID, allocatedRate, allocatedRate)
}

// canScheduleDownload checks if a host can be scheduled for download.
func (es *Distribute) canScheduleDownload(hostID string) bool {
	// logger.Infof("[distribute]: host %s idle download rate: %d, rate limit: %d", hostID, es.getHostIdleDownloadRate(hostID), es.rateLimit)
	return es.getHostIdleDownloadRate(hostID) > es.rateThreshold
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

	realRate := host.Network.UploadRate
	occupied := realRate
	allocatedRate := host.Network.AllocatedUploadRate
	if realRate < allocatedRate {
		occupied = allocatedRate
	}
	if host.Network.UploadRateLimit < occupied {
		return 0
	}

	return host.Network.UploadRateLimit - occupied
}

// getHostIdleDownloadRate gets the idle download rate for a host.
func (es *Distribute) getHostIdleDownloadRate(hostID string) uint64 {
	host, exists := es.resource.HostManager().Load(hostID)
	if !exists {
		return 0
	}

	realRate := host.Network.DownloadRate
	occupied := realRate
	allocatedRate := host.Network.AllocatedDownloadRate
	if realRate < allocatedRate {
		occupied = allocatedRate
	}
	if host.Network.DownloadRateLimit < occupied {
		return 0
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
