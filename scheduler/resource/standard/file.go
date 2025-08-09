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

package standard

import (
	"fmt"
	"math"
	"sync"
	"time"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/container/set"
	nethttp "d7y.io/dragonfly/v2/pkg/net/http"
	"github.com/bits-and-blooms/bitset"
)

// File represents a file and its blocks bound to a specific task.
type File struct {
	// ID is file id.
	ID string

	// TaskID is task id bound to this file.
	TaskID string

	// URL is file url.
	URL string

	// PieceLength is the length of pieces in this file
	PieceLength uint

	// ContentLength is file content length.
	ContentLength int64

	// TotalPartitions is total partitions.
	TotalPartitions uint

	// PartitionLength is default partition length.
	PartitionLength uint64

	// Partitions maps partition number to Partition.
	Partitions *sync.Map

	// FinishedPartitions is the set of completed partition numbers.
	FinishedPartitions *bitset.BitSet

	// Finished is the flag indicating whether the file is finished.
	Finished bool

	// FinishedAt is the time when the file is finished.
	FinishedAt time.Time

	// ActivePeers is the set of active peer ids downloading blocks.
	ActivePeers set.SafeSet[string]

	// Host is the host that owns this file (needed to access peer states)
	Host *Host

	// Mutex for thread safety.
	mutex sync.RWMutex
}

// NewFile creates a new file instance
func NewFile(id, taskID, url string, pieceLength uint, contentLength int64, blockLength uint64) *File {
	totalBlocks := uint(math.Ceil(float64(contentLength) / float64(blockLength)))
	return &File{
		ID:                 id,
		TaskID:             taskID,
		URL:                url,
		PieceLength:        pieceLength,
		ContentLength:      contentLength,
		TotalPartitions:    totalBlocks,
		PartitionLength:    blockLength,
		Partitions:         &sync.Map{},
		FinishedPartitions: &bitset.BitSet{},
		ActivePeers:        set.NewSafeSet[string](),
		Host:               nil, // Will be set when file is stored to host
	}
}

// AssignPartitionToPeer
func AssignPartitionToPeer(task *Task, host *Host, peer *Peer) error {
	if peer.Range == nil {
		logger.Warnf("[distribute]: peer %s has no range", peer.ID)
		return nil
	}

	file_id := task.ID
	file, ok := host.LoadFile(file_id)
	if !ok {
		logger.Warnf("[distribute]: file %s not found", file_id)
		return fmt.Errorf("file %s not found", file_id)
	}

	blockNumber := file.CalculatePartitionNumber(peer.Range)
	if blockNumber == -1 {
		logger.Warnf("[distribute]: block number is -1")
		return nil
	}

	if _, err := file.AssignPartitionToPeer(blockNumber, peer.ID); err != nil {
		logger.Warnf("[distribute]: failed to assign block %d to peer %s: %v", blockNumber, peer.ID, err)
		return err
	}

	block, _ := file.GetPartition(blockNumber)
	if block != nil {
		peer.AllocatedParents.Add(block.ParentID)
	}

	logger.Infof("[distribute]: assigned parent %s to peer %s", block.ParentID, peer.ID)

	return nil
}

func FinishPartitionByPeer(task *Task, host *Host, peer *Peer) error {
	file, ok := host.LoadFile(task.ID)
	if !ok {
		logger.Warnf("[distribute]: file %s not found", task.ID)
		return fmt.Errorf("file %s not found", task.ID)
	}

	blockNumber := file.CalculatePartitionNumber(peer.Range)
	if blockNumber == -1 {
		logger.Warnf("[distribute]: block number is -1")
		return fmt.Errorf("block number is -1")
	}

	file.FinishPartition(blockNumber)

	return nil
}

func FinishFile(resource Resource, hostID string, taskID string, peerID string) {
	if host, ok := resource.HostManager().Load(hostID); ok {
		if file, ok := host.LoadFile(taskID); ok {
			if !file.IsFinished() {
				for i := int32(0); i < int32(file.TotalPartitions); i++ {
					if file.FinishedPartitions.Test(uint(i)) {
						continue
					}

					if _, err := file.AssignPartitionToPeer(i, peerID); err != nil {
						logger.Warnf("[distribute]: failed to assign block %d to seed peer %s: %s", i, peerID, err.Error())
						continue
					}

					if err := file.FinishPartitionBackToSource(i); err != nil {
						logger.Warnf("[distribute]: failed to complete block %d for seed peer %s: %s", i, peerID, err.Error())
					}
				}
			}
		}
	}
}

// CreatePartitions creates all blocks for this file
func (f *File) CreatePartitions() {
	for i := uint(0); i < f.TotalPartitions; i++ {
		offset := uint64(i) * f.PartitionLength
		length := f.PartitionLength

		// Adjust length for the last block
		if offset+length > uint64(f.ContentLength) {
			length = uint64(f.ContentLength) - offset
		}

		block := NewPartition(int32(i), offset, length, f.TaskID, f.PieceLength)
		f.Partitions.Store(int32(i), block)
	}
}

// GetPartition returns a partition by its number
func (f *File) GetPartition(partitionNumber int32) (*Partition, bool) {
	value, exists := f.Partitions.Load(partitionNumber)
	if !exists {
		return nil, false
	}
	return value.(*Partition), true
}

// getPartitionWithValidation 获取 block 并验证其存在性
func (f *File) getPartitionWithValidation(blockNumber int32) (*Partition, error) {
	block, exists := f.GetPartition(blockNumber)
	if !exists {
		return nil, fmt.Errorf("block %d not found", blockNumber)
	}
	return block, nil
}

// collectPartitions 通用的块收集函数
func (f *File) collectPartitions(filterFn func(*Partition) bool) []*Partition {
	var blocks []*Partition
	f.forEachPartition(func(block *Partition) {
		if filterFn == nil || filterFn(block) {
			blocks = append(blocks, block)
		}
	})
	return blocks
}

// forEachPartition 遍历所有块并执行操作
func (f *File) forEachPartition(fn func(*Partition)) {
	f.Partitions.Range(func(key, value interface{}) bool {
		block := value.(*Partition)
		fn(block)
		return true
	})
}

// AssignPartitionToPeer assigns a partition to a specific peer
func (f *File) AssignPartitionToPeer(partitionNumber int32, peerID string) (*Partition, error) {
	var block *Partition
	var err error

	block, err = f.getPartitionWithValidation(partitionNumber)
	if err != nil {
		return nil, err
	}

	if block.GetPeerID() != "" && block.GetPeerID() != peerID {
		err = fmt.Errorf("partition %d already assigned to peer %s", partitionNumber, block.GetPeerID())
		return nil, err
	}

	block.AssignToPeer(peerID)
	f.ActivePeers.Add(peerID)

	return block, err
}

// FinishPartition marks a partition as completed
func (f *File) FinishPartition(partitionNumber int32) error {
	var err error

	block, blockErr := f.getPartitionWithValidation(partitionNumber)
	if blockErr != nil {
		err = blockErr
		return err
	}

	block.State = PartitionStateCompleted
	f.FinishedPartitions.Set(uint(partitionNumber))

	return err
}

func (f *File) FinishPartitionBackToSource(partitionNumber int32) error {
	block, err := f.getPartitionWithValidation(partitionNumber)
	if err != nil {
		return err
	}

	peerID := block.GetPeerID()
	peer, ok := f.Host.LoadPeer(block.GetPeerID())
	if !ok {
		return fmt.Errorf("peer %s not found", peerID)
	}

	var finished bool = true
	pieceNumbers := block.GetPieceNumbers()
	for _, pieceNumber := range pieceNumbers {
		if !peer.FinishedPieces.Test(uint(pieceNumber)) {
			finished = false
			break
		}
	}

	if finished {
		f.FinishPartition(partitionNumber)
	}

	return nil
}

// GetMissingPartitions returns block numbers that are not completed yet
func (f *File) GetMissingPartitions() []int32 {
	var missing []int32

	for i := int32(0); i < int32(f.TotalPartitions); i++ {
		block, exists := f.GetPartition(i)
		if !exists {
			continue
		}
		if block.State == PartitionStateCompleted || block.State == PartitionStateDownload {
			continue
		}

		if f.Host != nil {
			if peer, loaded := f.Host.LoadPeer(block.GetPeerID()); loaded {
				if peer.FSM.Is(PeerStateSucceeded) {
					// Update the cached status with proper write lock
					block.State = PartitionStateCompleted
					f.FinishedPartitions.Set(uint(i))
					continue
				}
			}
		}
		missing = append(missing, i)
	}

	return missing
}

// IsPartitionFinished checks if a block is completed using cached status and peer state
func (f *File) IsPartitionFinished(blockNumber int32) bool {
	block, exists := f.GetPartition(blockNumber)
	if !exists || block.GetPeerID() == "" {
		return false
	}

	// Fast path: if block is already marked as completed, return true
	if block.State == PartitionStateCompleted {
		return true
	}

	// Slow path: check peer state only if block is not completed
	if f.Host != nil {
		if peer, loaded := f.Host.LoadPeer(block.GetPeerID()); loaded {
			if peer.FSM.Is(PeerStateSucceeded) {
				// Update the cached status with proper write lock
				block.State = PartitionStateCompleted
				f.FinishedPartitions.Set(uint(blockNumber))
				return true
			}
		}
	}

	return false
}

func (f *File) IsPartitionDownloading(blockNumber int32) bool {
	block, exists := f.GetPartition(blockNumber)
	if !exists || block.GetPeerID() == "" {
		return false
	}
	return block.State == PartitionStateDownload
}

// IsPartitionAssigned checks if a block is being downloaded by a specific peer
func (f *File) IsPartitionAssigned(blockNumber int32, peerID string) bool {
	block, exists := f.GetPartition(blockNumber)
	if !exists || block.GetPeerID() != peerID {
		return false
	}

	if f.Host != nil {
		if peer, loaded := f.Host.LoadPeer(block.GetPeerID()); loaded {
			if !peer.FSM.Is(PeerStateSucceeded) {
				return true
			}
		}
	}

	return false
}

// GetPartitionsByPeer returns all partitions assigned to a specific peer
func (f *File) GetPartitionsByPeer(peerID string) []*Partition {
	return f.collectPartitions(func(block *Partition) bool {
		return block.GetPeerID() == peerID
	})
}

// IsFinished checks if all partitions are completed
func (f *File) IsFinished() bool {
	if f.Finished {
		return true
	}

	var result bool
	result = f.FinishedPartitions.Count() == uint(f.TotalPartitions)
	if result {
		f.Finished = result
		f.FinishedAt = time.Now()
	}
	return result
}

// GetCompletionRatio returns the completion ratio (0.0 to 1.0)
func (f *File) GetCompletionRatio() float64 {
	var result float64
	if f.TotalPartitions == 0 {
		result = 0.0
	} else {
		result = float64(f.FinishedPartitions.Count()) / float64(f.TotalPartitions)
	}
	return result
}

// RemovePeer removes a peer and resets all its assigned partitions
func (f *File) RemovePeer(peerID string) {
	f.forEachPartition(func(block *Partition) {
		if block.GetPeerID() == peerID {
			block.PeerID = ""
			f.FinishedPartitions.Clear(uint(block.Number))
		}
	})
	f.ActivePeers.Delete(peerID)
}

// GetPartitionHTTPRange returns the HTTP range string for a partition
func (f *File) GetPartitionHTTPRange(partitionNumber int32) (string, error) {
	block, err := f.getPartitionWithValidation(partitionNumber)
	if err != nil {
		return "", err
	}
	return block.GetHTTPRange(), nil
}

// GetAllPartitions returns all partitions as a slice
func (f *File) GetAllPartitions() []*Partition {
	return f.collectPartitions(nil)
}

// CalculatePartitionNumber calculates which partition number corresponds to a given HTTP range
func (f *File) CalculatePartitionNumber(r *nethttp.Range) int32 {
	if r == nil || r.Start < 0 || r.Length <= 0 {
		return -1 // 返回 -1 表示无效
	}

	// 根据 range 的起始位置计算对应的 partition number
	blockNumber := int32(r.Start / int64(f.PartitionLength))

	// 确保 partition number 在有效范围内
	if blockNumber < 0 {
		blockNumber = 0
	}
	if blockNumber >= int32(f.TotalPartitions) {
		blockNumber = int32(f.TotalPartitions) - 1
	}

	return blockNumber
}
