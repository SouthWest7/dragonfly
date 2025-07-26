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

	// ContentLength is file content length.
	ContentLength uint

	// TotalBlocks is total blocks.
	TotalBlocks uint

	// BlockLength is default block length.
	BlockLength uint

	// Blocks maps block number to Block.
	Blocks *sync.Map

	// FinishedBlocks is the set of completed block numbers.
	FinishedBlocks *bitset.BitSet

	// ActivePeers is the set of active peer ids downloading blocks.
	ActivePeers set.SafeSet[string]

	// Host is the host that owns this file (needed to access peer states)
	Host *Host

	// Mutex for thread safety.
	mutex sync.RWMutex
}

// NewFile creates a new file instance
func NewFile(id, taskID, url string, contentLength uint, blockLength uint) *File {
	totalBlocks := uint(math.Ceil(float64(contentLength) / float64(blockLength)))
	return &File{
		ID:             id,
		TaskID:         taskID,
		URL:            url,
		ContentLength:  contentLength,
		TotalBlocks:    totalBlocks,
		BlockLength:    blockLength,
		Blocks:         &sync.Map{},
		FinishedBlocks: &bitset.BitSet{},
		ActivePeers:    set.NewSafeSet[string](),
		Host:           nil, // Will be set when file is stored to host
	}
}

// AssignBlockToPeer
func AssignBlockToPeer(task *Task, host *Host, peer *Peer) error {
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

	blockNumber := file.CalculateBlockNumber(peer.Range)
	if blockNumber == -1 {
		logger.Warnf("[distribute]: block number is -1")
		return nil
	}

	if _, err := file.AssignBlockToPeer(blockNumber, peer.ID); err != nil {
		logger.Warnf("[distribute]: failed to assign block %d to peer %s: %v", blockNumber, peer.ID, err)
		return err
	}

	block, _ := file.GetBlock(blockNumber)
	if block != nil {
		peer.AllocatedParents.Add(block.ParentID)
	}

	logger.Infof("[distribute]: assigned block %d to peer %s", blockNumber, peer.ID)

	return nil
}

// CreateBlocks creates all blocks for this file
func (f *File) CreateBlocks() {
	f.withWriteLock(func() {
		for i := uint(0); i < f.TotalBlocks; i++ {
			offset := uint(i) * f.BlockLength
			length := f.BlockLength

			// Adjust length for the last block
			if offset+length > f.ContentLength {
				length = f.ContentLength - offset
			}

			block := NewBlock(int32(i), offset, length, f.TaskID)
			f.Blocks.Store(int32(i), block)
		}
	})
}

// GetBlock returns a block by its number
func (f *File) GetBlock(blockNumber int32) (*Block, bool) {
	value, exists := f.Blocks.Load(blockNumber)
	if !exists {
		return nil, false
	}
	return value.(*Block), true
}

// withWriteLock 执行需要写锁的操作
func (f *File) withWriteLock(fn func()) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	fn()
}

// withReadLock 执行需要读锁的操作
func (f *File) withReadLock(fn func()) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	fn()
}

// getBlockWithValidation 获取 block 并验证其存在性
func (f *File) getBlockWithValidation(blockNumber int32) (*Block, error) {
	block, exists := f.GetBlock(blockNumber)
	if !exists {
		return nil, fmt.Errorf("block %d not found", blockNumber)
	}
	return block, nil
}

// collectBlocks 通用的块收集函数
func (f *File) collectBlocks(filterFn func(*Block) bool) []*Block {
	var blocks []*Block
	f.Blocks.Range(func(key, value interface{}) bool {
		block := value.(*Block)
		if filterFn == nil || filterFn(block) {
			blocks = append(blocks, block)
		}
		return true
	})
	return blocks
}

// forEachBlock 遍历所有块并执行操作
func (f *File) forEachBlock(fn func(*Block)) {
	f.Blocks.Range(func(key, value interface{}) bool {
		block := value.(*Block)
		fn(block)
		return true
	})
}

// AssignBlockToPeer assigns a block to a specific peer
func (f *File) AssignBlockToPeer(blockNumber int32, peerID string) (*Block, error) {
	var block *Block
	var err error

	f.withWriteLock(func() {
		block, err = f.getBlockWithValidation(blockNumber)
		if err != nil {
			return
		}

		if block.GetPeerID() != "" && block.GetPeerID() != peerID {
			err = fmt.Errorf("block %d already assigned to peer %s", blockNumber, block.GetPeerID())
			return
		}

		block.AssignToPeer(peerID)
		f.ActivePeers.Add(peerID)
	})

	return block, err
}

// CompleteBlock marks a block as completed
func (f *File) CompleteBlock(blockNumber int32) error {
	var err error

	f.withWriteLock(func() {
		block, blockErr := f.getBlockWithValidation(blockNumber)
		if blockErr != nil {
			err = blockErr
			return
		}

		block.State = BlockStateCompleted
		f.FinishedBlocks.Set(uint(blockNumber))
	})

	return err
}

// GetMissingBlocks returns block numbers that are not completed yet
func (f *File) GetMissingBlocks() []int32 {
	var missing []int32

	for i := int32(0); i < int32(f.TotalBlocks); i++ {
		block, exists := f.GetBlock(i)
		if !exists {
			continue
		}
		if block.State == BlockStateCompleted || block.State == BlockStateDownload {
			continue
		}

		if f.Host != nil {
			if peer, loaded := f.Host.LoadPeer(block.GetPeerID()); loaded {
				if peer.FSM.Is(PeerStateSucceeded) {
					// Update the cached status with proper write lock
					f.withWriteLock(func() {
						block.State = BlockStateCompleted
						f.FinishedBlocks.Set(uint(i))
					})
					continue
				}
			}
		}
		missing = append(missing, i)
	}

	return missing
}

// IsBlockFinished checks if a block is completed using cached status and peer state
func (f *File) IsBlockFinished(blockNumber int32) bool {
	block, exists := f.GetBlock(blockNumber)
	if !exists || block.GetPeerID() == "" {
		return false
	}

	// Fast path: if block is already marked as completed, return true
	if block.State == BlockStateCompleted {
		return true
	}

	// Slow path: check peer state only if block is not completed
	if f.Host != nil {
		if peer, loaded := f.Host.LoadPeer(block.GetPeerID()); loaded {
			if peer.FSM.Is(PeerStateSucceeded) {
				// Update the cached status with proper write lock
				f.withWriteLock(func() {
					block.State = BlockStateCompleted
					f.FinishedBlocks.Set(uint(blockNumber))
				})
				return true
			}
		}
	}

	return false
}

func (f *File) IsBlockDownloading(blockNumber int32) bool {
	block, exists := f.GetBlock(blockNumber)
	if !exists || block.GetPeerID() == "" {
		return false
	}
	return block.State == BlockStateDownload
}

// IsBlockAssigned checks if a block is being downloaded by a specific peer
func (f *File) IsBlockAssigned(blockNumber int32, peerID string) bool {
	block, exists := f.GetBlock(blockNumber)
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

// GetBlocksByPeer returns all blocks assigned to a specific peer
func (f *File) GetBlocksByPeer(peerID string) []*Block {
	return f.collectBlocks(func(block *Block) bool {
		return block.GetPeerID() == peerID
	})
}

// IsFinished checks if all blocks are completed
func (f *File) IsFinished() bool {
	var result bool
	f.withReadLock(func() {
		result = f.FinishedBlocks.Count() == uint(f.TotalBlocks)
	})
	return result
}

// GetCompletionRatio returns the completion ratio (0.0 to 1.0)
func (f *File) GetCompletionRatio() float64 {
	var result float64
	f.withReadLock(func() {
		if f.TotalBlocks == 0 {
			result = 0.0
		} else {
			result = float64(f.FinishedBlocks.Count()) / float64(f.TotalBlocks)
		}
	})
	return result
}

// RemovePeer removes a peer and resets all its assigned blocks
func (f *File) RemovePeer(peerID string) {
	f.withWriteLock(func() {
		f.forEachBlock(func(block *Block) {
			if block.GetPeerID() == peerID {
				block.PeerID = ""
				f.FinishedBlocks.Clear(uint(block.Number))
			}
		})
		f.ActivePeers.Delete(peerID)
	})
}

// GetBlockHTTPRange returns the HTTP range string for a block
func (f *File) GetBlockHTTPRange(blockNumber int32) (string, error) {
	block, err := f.getBlockWithValidation(blockNumber)
	if err != nil {
		return "", err
	}
	return block.GetHTTPRange(), nil
}

// GetAllBlocks returns all blocks as a slice
func (f *File) GetAllBlocks() []*Block {
	return f.collectBlocks(nil)
}

// CalculateBlockNumber calculates which block number corresponds to a given HTTP range
func (f *File) CalculateBlockNumber(r *nethttp.Range) int32 {
	if r == nil || r.Start < 0 || r.Length <= 0 {
		return -1 // 返回 -1 表示无效
	}

	// 根据 range 的起始位置计算对应的 block number
	blockNumber := int32(r.Start / int64(f.BlockLength))

	// 确保 block number 在有效范围内
	if blockNumber < 0 {
		blockNumber = 0
	}
	if blockNumber >= int32(f.TotalBlocks) {
		blockNumber = int32(f.TotalBlocks) - 1
	}

	return blockNumber
}
