/*
 *     Copyright 2020 The Dragonfly Authors
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
	"time"
)

// PreheatRequest defines the request parameters for preheating.
type PreheatRequest struct {
	URL                         string            `json:"url" validate:"required,url"`
	PieceLength                 *uint64           `json:"pieceLength" binding:"omitempty,gte=4194304"`
	Tag                         string            `json:"tag" validate:"omitempty"`
	FilteredQueryParams         string            `json:"filtered_query_params" validate:"omitempty"`
	Headers                     map[string]string `json:"headers" validate:"omitempty"`
	Application                 string            `json:"application" validate:"omitempty"`
	Priority                    int32             `json:"priority" validate:"omitempty"`
	Scope                       string            `json:"scope" validate:"omitempty"`
	Percentage                  *uint8            `json:"percentage" validate:"omitempty,gte=1,lte=100"`
	Count                       *uint32           `json:"count" validate:"omitempty,gte=1,lte=200"`
	ConcurrentCount             int64             `json:"concurrent_count" validate:"omitempty"`
	CertificateChain            [][]byte          `json:"certificate_chain" validate:"omitempty"`
	InsecureSkipVerify          bool              `json:"insecure_skip_verify" validate:"omitempty"`
	Timeout                     time.Duration     `json:"timeout" validate:"omitempty"`
	LoadToCache                 bool              `json:"load_to_cache" validate:"omitempty"`
	ContentForCalculatingTaskID *string           `json:"content_for_calculating_task_id" validate:"omitempty"`
}

// PreheatResponse defines the response parameters for preheating.
type PreheatResponse struct {
	SuccessTasks       []*PreheatSuccessTask `json:"success_tasks"`
	FailureTasks       []*PreheatFailureTask `json:"failure_tasks"`
	SchedulerClusterID uint                  `json:"scheduler_cluster_id"`
}

// PreheatSuccessTask defines the response parameters for preheating successfully.
type PreheatSuccessTask struct {
	URL      string `json:"url"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// PreheatFailureTask defines the response parameters for preheating failed.
type PreheatFailureTask struct {
	URL         string `json:"url"`
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	Description string `json:"description"`
}

// GetTaskRequest defines the request parameters for getting task.
type GetTaskRequest struct {
	TaskID  string        `json:"task_id" validate:"required"`
	Timeout time.Duration `json:"timeout" validate:"omitempty"`
}

// GetTaskResponse defines the response parameters for getting task.
type GetTaskResponse struct {
	Peers              []*Peer `json:"peers"`
	SchedulerClusterID uint    `json:"scheduler_cluster_id"`
}

// Peer represents the peer information.
type Peer struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	HostType  string    `json:"host_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeleteTaskRequest defines the request parameters for deleting task.
type DeleteTaskRequest struct {
	TaskID  string        `json:"task_id" validate:"required"`
	Timeout time.Duration `json:"timeout" validate:"omitempty"`
}

// DeleteTaskResponse defines the response parameters for deleting task.
type DeleteTaskResponse struct {
	SuccessTasks       []*DeleteSuccessTask `json:"success_tasks"`
	FailureTasks       []*DeleteFailureTask `json:"failure_tasks"`
	SchedulerClusterID uint                 `json:"scheduler_cluster_id"`
}

// DeleteSuccessTask defines the response parameters for deleting peer successfully.
type DeleteSuccessTask struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	HostType string `json:"host_type"`
}

// DeleteFailureTask defines the response parameters for deleting peer failed.
type DeleteFailureTask struct {
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	HostType    string `json:"host_type"`
	Description string `json:"description"`
}

// DistributeRequest defines the request parameters for distribute.
type DistributeRequest struct {
	// URL is the download URL for the file.
	URL string `json:"url" validate:"required,url"`

	// PieceLength is the piece length for the file.
	PieceLength *uint64 `json:"piece_length" validate:"omitempty,gte=4194304"`

	// BlockLength is the block length for the file.
	BlockLength *uint64 `json:"block_length" validate:"omitempty,gte=4194304"`

	// ContentLength is the content length for the file.
	ContentLength *uint64 `json:"content_length" validate:"omitempty,gte=1"`

	// RateLimit is the rate limit for the distribute.
	RateLimit *uint64 `json:"rate_limit" validate:"omitempty,gte=1048576"`

	// ScheduleInterval is the schedule interval for the distribute.
	ScheduleInterval *uint64 `json:"schedule_interval" validate:"omitempty,gte=1"`

	// Tag is the tag for distribute.
	Tag string `json:"tag" validate:"omitempty"`

	// Application is the application string for distribute.
	Application string `json:"application" validate:"omitempty"`

	// FilteredQueryParams is the filtered query params for distribute.
	FilteredQueryParams string `json:"filtered_query_params" validate:"omitempty"`

	// Headers is the http headers for authentication.
	Headers map[string]string `json:"headers" validate:"omitempty"`
}

// DistributeResponse defines the response parameters for distribute.
type DistributeResponse struct {
	// SuccessTasks is the list of successful distribute tasks.
	SuccessTasks []*DistributeSuccessTask `json:"success_tasks"`

	// FailureTasks is the list of failed distribute tasks.
	FailureTasks []*DistributeFailureTask `json:"failure_tasks"`

	// SchedulerClusterID is the ID of the scheduler cluster.
	SchedulerClusterID uint `json:"scheduler_cluster_id"`

	// TotalScheduled is the total number of scheduled downloads.
	TotalScheduled int `json:"total_scheduled"`

	// ActiveDownloads is the number of active downloads.
	ActiveDownloads int `json:"active_downloads"`
}

// DistributeSuccessTask defines the response parameters for successful distribute.
type DistributeSuccessTask struct {
	// URL is the download URL.
	URL string `json:"url"`

	// Hostname is the hostname of the target host.
	Hostname string `json:"hostname"`

	// IP is the IP address of the target host.
	IP string `json:"ip"`

	// TaskID is the ID of the download task.
	TaskID string `json:"task_id"`

	// BlockID is the ID of the block being downloaded.
	BlockID int `json:"block_id"`

	// Rate is the download rate in bytes per second.
	Rate uint64 `json:"rate"`

	// EstimatedEndTime is the estimated end time of the download.
	EstimatedEndTime time.Time `json:"estimated_end_time"`
}

// DistributeFailureTask defines the response parameters for failed distribute.
type DistributeFailureTask struct {
	// URL is the download URL.
	URL string `json:"url"`

	// Hostname is the hostname of the target host.
	Hostname string `json:"hostname"`

	// IP is the IP address of the target host.
	IP string `json:"ip"`

	// Description is the error description.
	Description string `json:"description"`
}
