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
	"time"
)

// BlockState is the state of a block
type BlockState string

const (
	BlockStatePending   BlockState = "pending"
	BlockStateDownload  BlockState = "download"
	BlockStateCompleted BlockState = "completed"
	BlockStateFailed    BlockState = "failed"
)

// Block represents a file block with its metadata
type Block struct {
	// Number is the block number/id
	Number int32

	// Offset is the start offset of this block in the file
	Offset uint64

	// Length is the length of this block in bytes
	Length uint64

	// PieceLength is the length of pieces
	PieceLength uint

	// FileID is the file id associated with this block
	FileID string

	// TaskID is the task id associated with this block
	TaskID string

	// PeerID is the peer id that is downloading/has downloaded this block
	PeerID string

	// ParentID is the parent id that is uploading this block
	ParentID string

	// CreatedAt is the time when block was created
	CreatedAt time.Time

	// UpdatedAt is the time when block was last updated
	UpdatedAt time.Time

	// State is the state of the block
	State BlockState
}

// NewBlock creates a new block instance
func NewBlock(number int32, offset, length uint64, taskID string, pieceLength uint) *Block {
	return &Block{
		Number:      number,
		Offset:      offset,
		Length:      length,
		PieceLength: pieceLength,
		TaskID:      taskID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       BlockStatePending,
	}
}

// GetHTTPRange returns the HTTP range string for this block
func (b *Block) GetHTTPRange() string {
	endOffset := b.Offset + b.Length - 1
	return fmt.Sprintf("bytes=%d-%d", b.Offset, endOffset)
}

// GetEndOffset returns the end offset of this block
func (b *Block) GetEndOffset() uint64 {
	return b.Offset + b.Length - 1
}

func (b *Block) GetPeerID() string {
	return b.PeerID
}

// AssignToPeer assigns this block to a specific peer
func (b *Block) AssignToPeer(peerID string) {
	b.PeerID = peerID
	b.UpdatedAt = time.Now()
}

// Reset resets the block state (removes peer assignment and completion status)
func (b *Block) Reset() {
	b.PeerID = ""
	b.State = BlockStatePending
	b.UpdatedAt = time.Now()
}

func (b *Block) GetPieceNumbers() []uint32 {
	if b.Length == 0 {
		return []uint32{}
	}

	startPieceNum := uint32(b.Offset / uint64(b.PieceLength))
	endPieceNum := uint32((b.Offset + b.Length - 1) / uint64(b.PieceLength))

	pieceCount := endPieceNum - startPieceNum + 1
	pieceNumbers := make([]uint32, 0, pieceCount)

	for pieceNum := startPieceNum; pieceNum <= endPieceNum; pieceNum++ {
		pieceNumbers = append(pieceNumbers, pieceNum)
	}

	return pieceNumbers
}
