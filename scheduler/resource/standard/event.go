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
 * See the License for the specific governing permissions and
 * limitations under the License.
 */

package standard

import (
	"strconv"
)

// Event represents a download event in the scheduler
type Event struct {
	// ID is the ID of the event.
	ID string

	// BlockNumber is the number of the block.
	BlockNumber int32

	// TaskID is the ID of the task.
	TaskID string

	// DownloadPeer is the peer that will download the block.
	DownloadPeer string

	// UploadPeers is the list of peers that will upload the block.
	UploadPeers []string

	// AllocatedUploadRate is the allocated upload rate of the event.
	AllocatedUploadRate uint64

	// AllocatedDownloadRate is the allocated download rate of the event.
	AllocatedDownloadRate uint64
}

// NewEvent creates a new event
func NewEvent(id string, blockNumber int32, taskID string, hostID string, parentID string, allocatedUploadRate uint64, allocatedDownloadRate uint64) *Event {
	return &Event{
		ID:                    id,
		BlockNumber:           blockNumber,
		TaskID:                taskID,
		DownloadPeer:          hostID,
		UploadPeers:           []string{parentID},
		AllocatedUploadRate:   allocatedUploadRate,
		AllocatedDownloadRate: allocatedDownloadRate,
	}
}

func EventID(hostID string, fileID string, blockNumber int32) string {
	return hostID + "-" + fileID + "-" + strconv.Itoa(int(blockNumber))
}
