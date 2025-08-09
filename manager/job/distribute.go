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
	"context"
	"fmt"
	"time"

	machineryv1tasks "github.com/dragonflyoss/machinery/v1/tasks"
	"github.com/google/uuid"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	internaljob "d7y.io/dragonfly/v2/internal/job"
	"d7y.io/dragonfly/v2/manager/models"
	"d7y.io/dragonfly/v2/manager/types"
)

// Distribute is an interface for distribute job.
type Distribute interface {
	// CreateDistribute creates a distribute job.
	CreateDistribute(context.Context, []models.Scheduler, types.DistributeArgs) (*internaljob.GroupJobState, error)
}

// distribute is an implementation of Distribute.
type distribute struct {
	job *internaljob.Job
}

// newDistribute creates a new Distribute.
func newDistribute(job *internaljob.Job) Distribute {
	return &distribute{
		job: job,
	}
}

// CreateDistribute creates a distribute job.
func (cs *distribute) CreateDistribute(ctx context.Context, schedulers []models.Scheduler, args types.DistributeArgs) (*internaljob.GroupJobState, error) {
	// Convert args to internal distribute request
	req := &internaljob.DistributeRequest{
		URL:                 args.URL,
		PieceLength:         args.PieceLength,
		BlockLength:         args.BlockLength,
		InitPeers:           args.InitPeers,
		RateThreshold:       args.RateThreshold,
		Usage:               args.Usage,
		ScheduleInterval:    args.ScheduleInterval,
		Tag:                 args.Tag,
		Application:         args.Application,
		FilteredQueryParams: args.FilteredQueryParams,
		Headers:             args.Headers,
	}

	// Get scheduler queues
	queues, err := getSchedulerQueues(schedulers)
	if err != nil {
		return nil, err
	}

	// Create group job
	return cs.createGroupJob(ctx, req, queues)
}

// createGroupJob creates a group job for distribute.
func (cs *distribute) createGroupJob(ctx context.Context, req *internaljob.DistributeRequest, queues []internaljob.Queue) (*internaljob.GroupJobState, error) {
	var signatures []*machineryv1tasks.Signature
	for _, queue := range queues {
		args, err := internaljob.MarshalRequest(req)
		if err != nil {
			logger.Errorf("[distribute]: marshal request error: %v", err)
			continue
		}

		signatures = append(signatures, &machineryv1tasks.Signature{
			UUID:       fmt.Sprintf("task_%s", uuid.New().String()),
			Name:       internaljob.DistributeJob,
			RoutingKey: queue.String(),
			Args:       args,
		})
	}

	group, err := machineryv1tasks.NewGroup(signatures...)
	if err != nil {
		return nil, err
	}

	var tasks []machineryv1tasks.Signature
	for _, signature := range signatures {
		tasks = append(tasks, *signature)
	}

	logger.Infof("[distribute]: create distribute group %s in queues %v, tasks: %#v", group.GroupUUID, queues, tasks)
	if _, err := cs.job.Server.SendGroupWithContext(ctx, group, 50); err != nil {
		logger.Errorf("[distribute]: create distribute group %s failed: %v", group.GroupUUID, err)
		return nil, err
	}

	return &internaljob.GroupJobState{
		GroupUUID: group.GroupUUID,
		State:     machineryv1tasks.StatePending,
		CreatedAt: time.Now(),
	}, nil
}
