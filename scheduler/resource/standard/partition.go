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

// PartitionState is the state of a partition
type PartitionState string

const (
	PartitionStatePending   PartitionState = "pending"
	PartitionStateDownload  PartitionState = "download"
	PartitionStateCompleted PartitionState = "completed"
	PartitionStateFailed    PartitionState = "failed"
)

// Partition represents a file partition with its metadata
type Partition struct {
	// Number is the partition number/id
	Number int32

	// Offset is the start offset of this partition in the file
	Offset uint64

	// Length is the length of this partition in bytes
	Length uint64

	// PieceLength is the length of pieces
	PieceLength uint

	// FileID is the file id associated with this partition
	FileID string

	// TaskID is the task id associated with this partition
	TaskID string

	// PeerID is the peer id that is downloading/has downloaded this partition
	PeerID string

	// ParentID is the parent id that is uploading this partition
	ParentID string

	// CreatedAt is the time when partition was created
	CreatedAt time.Time

	// UpdatedAt is the time when partition was last updated
	UpdatedAt time.Time

	// State is the state of the partition
	State PartitionState
}

// NewPartition creates a new partition instance
func NewPartition(number int32, offset, length uint64, taskID string, pieceLength uint) *Partition {
	return &Partition{
		Number:      number,
		Offset:      offset,
		Length:      length,
		PieceLength: pieceLength,
		TaskID:      taskID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		State:       PartitionStatePending,
	}
}

// GetHTTPRange returns the HTTP range string for this partition
func (p *Partition) GetHTTPRange() string {
	endOffset := p.Offset + p.Length - 1
	return fmt.Sprintf("bytes=%d-%d", p.Offset, endOffset)
}

// GetEndOffset returns the end offset of this partition
func (p *Partition) GetEndOffset() uint64 {
	return p.Offset + p.Length - 1
}

func (p *Partition) GetPeerID() string {
	return p.PeerID
}

// AssignToPeer assigns this partition to a specific peer
func (p *Partition) AssignToPeer(peerID string) {
	p.PeerID = peerID
	p.UpdatedAt = time.Now()
}

// Reset resets the partition state (removes peer assignment and completion status)
func (p *Partition) Reset() {
	p.PeerID = ""
	p.State = PartitionStatePending
	p.UpdatedAt = time.Now()
}

func (p *Partition) GetPieceCount() uint32 {
	if p.Length == 0 {
		return 0
	}

	count := uint32(p.Length / uint64(p.PieceLength))
	if count*uint32(p.PieceLength) < uint32(p.Length) {
		return count + 1
	}

	return count
}

func (p *Partition) GetPieceNumbers() []uint32 {
	pieceCount := p.GetPieceCount()
	if pieceCount == 0 {
		return []uint32{}
	}

	startPieceNum := uint32(p.Offset / uint64(p.PieceLength))
	pieceNumbers := make([]uint32, 0, pieceCount)

	for i := uint32(0); i < pieceCount; i++ {
		pieceNumbers = append(pieceNumbers, startPieceNum+i)
	}

	return pieceNumbers
}
