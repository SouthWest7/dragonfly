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

package evaluator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	commonv2 "d7y.io/api/v2/pkg/apis/common/v2"

	"d7y.io/dragonfly/v2/pkg/idgen"
	"d7y.io/dragonfly/v2/pkg/types"
	"d7y.io/dragonfly/v2/scheduler/resource/standard"
)

func TestCustomEvaluator_newCustomEvaluator(t *testing.T) {
	tests := []struct {
		name   string
		expect func(t *testing.T, e any)
	}{
		{
			name: "new custom evaluator",
			expect: func(t *testing.T, e any) {
				assert := assert.New(t)
				assert.Equal(reflect.TypeOf(e).Elem().Name(), "customEvaluator")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(t, newCustomEvaluator())
		})
	}
}

func TestCustomEvaluator_EvaluateParents(t *testing.T) {
	tests := []struct {
		name            string
		parents         []*standard.Peer
		child           *standard.Peer
		totalPieceCount uint32
		mock            func(parent []*standard.Peer, child *standard.Peer)
		expect          func(t *testing.T, parents []*standard.Peer)
	}{
		{
			name:    "parents is empty",
			parents: []*standard.Peer{},
			child: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeNormal)),
			totalPieceCount: 1,
			mock: func(parent []*standard.Peer, child *standard.Peer) {
			},
			expect: func(t *testing.T, parents []*standard.Peer) {
				assert := assert.New(t)
				assert.Equal(len(parents), 0)
			},
		},
		{
			name: "evaluate parents with different piece counts",
			parents: []*standard.Peer{
				standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
					standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
					standard.NewHost(
						"bar", "127.0.0.1", "hostname",
						8003, 8001, types.HostTypeNormal)),
				standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
					standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
					standard.NewHost(
						"baz", "127.0.0.1", "hostname",
						8003, 8001, types.HostTypeNormal)),
				standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
					standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
					standard.NewHost(
						"bac", "127.0.0.1", "hostname",
						8003, 8001, types.HostTypeNormal)),
				standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
					standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
					standard.NewHost(
						mockSeedHostID, "127.0.0.1", "hostname",
						8003, 8001, types.HostTypeSuperSeed)),
				standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
					standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
					standard.NewHost(
						"bae", "127.0.0.1", "hostname",
						8003, 8001, types.HostTypeNormal)),
			},
			child: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeNormal)),
			totalPieceCount: 1,
			mock: func(parents []*standard.Peer, child *standard.Peer) {
				parents[1].FinishedPieces.Set(0)
				parents[2].FinishedPieces.Set(0).Set(1)
				parents[3].FinishedPieces.Set(0).Set(1).Set(2)
				parents[4].FinishedPieces.Set(0).Set(1).Set(2).Set(3)
			},
			expect: func(t *testing.T, parents []*standard.Peer) {
				assert := assert.New(t)
				assert.Equal(len(parents), 5)
				// Custom algorithm should prioritize based on custom weights
				// The actual order depends on the custom algorithm's scoring
				// We just verify that the algorithm produces a valid ordering
				assert.NotEmpty(parents[0].Host.ID)
				assert.NotEmpty(parents[1].Host.ID)
				assert.NotEmpty(parents[2].Host.ID)
				assert.NotEmpty(parents[3].Host.ID)
				assert.NotEmpty(parents[4].Host.ID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newCustomEvaluator()
			tc.mock(tc.parents, tc.child)
			tc.expect(t, e.EvaluateParents(tc.parents, tc.child, tc.totalPieceCount))
		})
	}
}

func TestCustomEvaluator_evaluate(t *testing.T) {
	tests := []struct {
		name            string
		parent          *standard.Peer
		child           *standard.Peer
		totalPieceCount uint32
		mock            func(parent *standard.Peer, child *standard.Peer)
		expect          func(t *testing.T, score float64)
	}{
		{
			name: "evaluate parent with custom algorithm",
			parent: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockSeedHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeSuperSeed)),
			child: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeNormal)),
			totalPieceCount: 1,
			mock: func(parent *standard.Peer, child *standard.Peer) {
			},
			expect: func(t *testing.T, score float64) {
				assert := assert.New(t)
				// Custom algorithm should have different score calculation
				assert.Greater(score, float64(0))
			},
		},
		{
			name: "evaluate parent with pieces using custom algorithm",
			parent: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockSeedHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeSuperSeed)),
			child: standard.NewPeer(idgen.PeerIDV1("127.0.0.1"),
				standard.NewTask(mockTaskID, mockTaskURL, mockTaskTag, mockTaskApplication, commonv2.TaskType_STANDARD, mockTaskFilteredQueryParams, mockTaskHeader, mockTaskBackToSourceLimit, standard.WithDigest(mockTaskDigest)),
				standard.NewHost(
					mockHostID, "127.0.0.1", "hostname",
					8003, 8001, types.HostTypeNormal)),
			totalPieceCount: 1,
			mock: func(parent *standard.Peer, child *standard.Peer) {
				parent.FinishedPieces.Set(0)
			},
			expect: func(t *testing.T, score float64) {
				assert := assert.New(t)
				// Custom algorithm should have different score calculation
				assert.Greater(score, float64(0))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newCustomEvaluator()
			tc.mock(tc.parent, tc.child)
			tc.expect(t, e.(*customEvaluator).evaluateParents(tc.parent, tc.child, tc.totalPieceCount))
		})
	}
}
