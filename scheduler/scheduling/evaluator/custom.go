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

package evaluator

import (
	"context"
	"math"
	"sort"
	"time"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/scheduler/resource/persistentcache"
	"d7y.io/dragonfly/v2/scheduler/resource/standard"
)

const (
	// DefaultPieceSize is the default piece size (32MB)
	DefaultPieceSize = 32 * 1024 * 1024
)

// customEvaluator is evaluator for custom evaluator.
type customEvaluator struct {
	// No need for additional fields - use existing File and Block structures
}

// newCustomEvaluator returns a new custom evaluator.
func newCustomEvaluator() Evaluator {
	return &customEvaluator{}
}

// EvaluateParents sort parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) EvaluateParents(parents []*standard.Peer, child *standard.Peer, totalPieceCount uint32) []*standard.Peer {
	// Cold start: Process peer range and bind to blocks using File's built-in methods
	if child.Range != nil {
		logger.Infof("Cold start: Peer %s with HTTP range: start=%d, length=%d (end=%d)",
			child.ID, child.Range.Start, child.Range.Length, child.Range.Start+child.Range.Length-1)

		// Calculate required blocks for this peer
		requiredBlocks := e.calculateRequiredBlocks(child)

		// Use existing File methods to assign blocks
		if child.Host != nil {
			if file, found := child.Host.LoadFile(child.Task.ID); found {
				for _, blockNum := range requiredBlocks {
					if _, err := file.AssignBlockToPeer(blockNum, child.ID); err != nil {
						logger.Warnf("Failed to assign block %d to peer %s: %s", blockNum, child.ID, err.Error())
					}
				}
				logger.Infof("Cold start: Peer %s assigned to %d blocks", child.ID, len(requiredBlocks))
			}
		}
	} else {
		logger.Infof("Cold start: Peer %s without range (full file download)", child.ID)
	}

	// Calculate required blocks for child
	requiredBlocks := e.calculateRequiredBlocks(child)

	// Filter and sort parents based on block availability
	filteredParents := e.filterParentsByBlockAvailability(parents, requiredBlocks)

	// Sort by block matching score first, then by finished pieces count
	sort.Slice(filteredParents, func(i, j int) bool {
		// Calculate block matching scores
		matchScoreI := e.calculateBlockMatchScore(filteredParents[i], requiredBlocks)
		matchScoreJ := e.calculateBlockMatchScore(filteredParents[j], requiredBlocks)

		// If block matching scores are equal, compare by finished pieces count
		if matchScoreI == matchScoreJ {
			return filteredParents[i].FinishedPieces.Count() > filteredParents[j].FinishedPieces.Count()
		}

		return matchScoreI > matchScoreJ
	})

	// Log cold start scheduling result
	logger.Infof("Cold start: Child peer %s needs %d blocks, found %d suitable parents",
		child.ID, len(requiredBlocks), len(filteredParents))

	return filteredParents
}

// calculateRequiredBlocks calculates which blocks are required for a child peer
func (e *customEvaluator) calculateRequiredBlocks(child *standard.Peer) []int32 {
	if child.Range == nil {
		// If no range specified, child needs all blocks
		totalBlocks := int32(math.Ceil(float64(child.Task.ContentLength.Load()) / float64(DefaultPieceSize)))
		var allBlocks []int32
		for i := int32(0); i < totalBlocks; i++ {
			allBlocks = append(allBlocks, i)
		}
		return allBlocks
	}

	return e.calculateBlockNumbers(child.Range.Start, child.Range.Length, DefaultPieceSize)
}

// calculateBlockNumbers calculates which block numbers a range covers
func (e *customEvaluator) calculateBlockNumbers(start, length, pieceSize int64) []int32 {
	startBlock := int32(start / pieceSize)
	endBlock := int32((start + length - 1) / pieceSize)

	var blocks []int32
	for i := startBlock; i <= endBlock; i++ {
		blocks = append(blocks, i)
	}

	return blocks
}

// filterParentsByBlockAvailability filters parents that have required blocks
func (e *customEvaluator) filterParentsByBlockAvailability(parents []*standard.Peer, requiredBlocks []int32) []*standard.Peer {
	var filteredParents []*standard.Peer

	for _, parent := range parents {
		if e.hasRequiredBlocks(parent, requiredBlocks) {
			filteredParents = append(filteredParents, parent)
		}
	}

	return filteredParents
}

// hasRequiredBlocks checks if a parent has any of the required blocks
func (e *customEvaluator) hasRequiredBlocks(parent *standard.Peer, requiredBlocks []int32) bool {
	// Check if parent has finished pieces that overlap with required blocks
	for _, blockNum := range requiredBlocks {
		if parent.FinishedPieces.Test(uint(blockNum)) {
			return true
		}
	}

	return false
}

// calculateBlockMatchScore calculates how well a parent matches required blocks
func (e *customEvaluator) calculateBlockMatchScore(parent *standard.Peer, requiredBlocks []int32) float64 {
	if len(requiredBlocks) == 0 {
		return 0.0
	}

	matchedBlocks := 0
	for _, blockNum := range requiredBlocks {
		if parent.FinishedPieces.Test(uint(blockNum)) {
			matchedBlocks++
		}
	}

	return float64(matchedBlocks) / float64(len(requiredBlocks))
}

// HandleEventSchedulerDownload handles download requests from event scheduler
func (e *customEvaluator) HandleEventSchedulerDownload(peer *standard.Peer, blockNumber int32) error {
	if peer.Host != nil {
		if file, found := peer.Host.LoadFile(peer.Task.ID); found {
			if _, err := file.AssignBlockToPeer(blockNumber, peer.ID); err != nil {
				return err
			}
			logger.Infof("Event scheduler: Bound peer %s to block %d for task %s",
				peer.ID, blockNumber, peer.Task.ID)
		}
	}
	return nil
}

// AssignBlockToPeer assigns a specific block to a peer (for event scheduler)
func (e *customEvaluator) AssignBlockToPeer(peer *standard.Peer, blockNumber int32) error {
	if peer.Host != nil {
		if file, found := peer.Host.LoadFile(peer.Task.ID); found {
			if _, err := file.AssignBlockToPeer(blockNumber, peer.ID); err != nil {
				return err
			}
			logger.Infof("Event scheduler: Assigned block %d to peer %s (task %s)",
				blockNumber, peer.ID, peer.Task.ID)
		}
	}
	return nil
}

// SelectBlockForPeer selects the best block for a peer using custom logic
func (e *customEvaluator) SelectBlockForPeer(peer *standard.Peer, availableBlocks []int32) int32 {
	if len(availableBlocks) == 0 {
		return -1
	}

	// Get required blocks for this peer
	requiredBlocks := e.calculateRequiredBlocks(peer)
	if len(requiredBlocks) == 0 {
		return availableBlocks[0] // Return first available block
	}

	// Find intersection of required blocks and available blocks
	for _, requiredBlock := range requiredBlocks {
		for _, availableBlock := range availableBlocks {
			if requiredBlock == availableBlock {
				return requiredBlock
			}
		}
	}

	// If no intersection, return first available block
	return availableBlocks[0]
}

// OnPeerCompleted handles peer completion events (can be called externally)
func (e *customEvaluator) OnPeerCompleted(peer *standard.Peer) {
	// Use existing File methods to handle completion
	if peer.Host != nil {
		if file, found := peer.Host.LoadFile(peer.Task.ID); found {
			// Get blocks assigned to this peer and mark them as completed
			blocks := file.GetBlocksByPeer(peer.ID)
			for _, block := range blocks {
				if err := file.CompleteBlock(block.Number, peer.ID); err != nil {
					logger.Warnf("Failed to complete block %d for peer %s: %s", block.Number, peer.ID, err.Error())
				}
			}

			// Log peer completion with its range info
			if peer.Range != nil {
				logger.Infof("Peer %s completed (range: %d-%d), task %s",
					peer.ID, peer.Range.Start, peer.Range.Start+peer.Range.Length-1, peer.Task.ID)
			} else {
				logger.Infof("Peer %s completed (full file), task %s", peer.ID, peer.Task.ID)
			}

			// Log completion progress
			completionRatio := file.GetCompletionRatio()
			logger.Infof("Task %s progress: %.1f%% completed", peer.Task.ID, completionRatio*100)
		}
	}
}

// StartMonitoring starts block monitoring
func (e *customEvaluator) StartMonitoring(ctx context.Context) {
	go e.monitorBlocks(ctx)
}

// monitorBlocks monitors block download progress
func (e *customEvaluator) monitorBlocks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.logBlockProgress()
		}
	}
}

// logBlockProgress logs the progress of block downloads
func (e *customEvaluator) logBlockProgress() {
	// This would need to be implemented based on how we track tasks
	// For now, we'll skip this as it requires access to the task manager
	logger.Infof("Block progress monitoring (simplified)")
}

// IsBadParent determine if peer is a bad parent, it can not be selected as a parent.
func (e *customEvaluator) IsBadParent(peer *standard.Peer) bool {
	return false
}

// EvaluatePersistentCacheParents sort persistent cache parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) EvaluatePersistentCacheParents(parents []*persistentcache.Peer, child *persistentcache.Peer, totalPieceCount uint32) []*persistentcache.Peer {
	// Simple sort by finished pieces count
	sort.Slice(parents, func(i, j int) bool {
		return parents[i].FinishedPieces.Count() > parents[j].FinishedPieces.Count()
	})

	return parents
}

// IsBadPersistentCacheParent determine if persistent cache peer is a bad parent, it can not be selected as a parent.
func (e *customEvaluator) IsBadPersistentCacheParent(peer *persistentcache.Peer) bool {
	return false
}
