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
	"sort"
	"strings"

	"d7y.io/dragonfly/v2/pkg/types"
	"d7y.io/dragonfly/v2/scheduler/resource/persistentcache"
	"d7y.io/dragonfly/v2/scheduler/resource/standard"
)

const (
	// Custom algorithm weights
	customFinishedPieceWeight        = 0.25
	customParentHostUploadSuccessWeight = 0.20
	customFreeUploadWeight           = 0.20
	customHostTypeWeight             = 0.15
	customIDCAffinityWeight          = 0.10
	customLocationAffinityWeight     = 0.10
)

// customEvaluator is a custom implementation of Evaluator.
type customEvaluator struct {
	evaluator
}

// newCustomEvaluator returns a new CustomEvaluator.
func newCustomEvaluator() Evaluator {
	return &customEvaluator{}
}

// EvaluateParents sort parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) EvaluateParents(parents []*standard.Peer, child *standard.Peer, totalPieceCount uint32) []*standard.Peer {
	sort.Slice(
		parents,
		func(i, j int) bool {
			return e.evaluateParents(parents[i], child, totalPieceCount) > e.evaluateParents(parents[j], child, totalPieceCount)
		},
	)

	return parents
}

// evaluateParents sort parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) evaluateParents(parent *standard.Peer, child *standard.Peer, totalPieceCount uint32) float64 {
	parentLocation := parent.Host.Network.Location
	parentIDC := parent.Host.Network.IDC
	childLocation := child.Host.Network.Location
	childIDC := child.Host.Network.IDC

	return customFinishedPieceWeight*e.calculatePieceScore(parent.FinishedPieces.Count(), child.FinishedPieces.Count(), totalPieceCount) +
		customParentHostUploadSuccessWeight*e.calculateParentHostUploadSuccessScore(parent.Host.UploadCount.Load(), parent.Host.UploadFailedCount.Load()) +
		customFreeUploadWeight*e.calculateFreeUploadScore(parent.Host) +
		customHostTypeWeight*e.calculateHostTypeScore(parent) +
		customIDCAffinityWeight*e.calculateIDCAffinityScore(parentIDC, childIDC) +
		customLocationAffinityWeight*e.calculateMultiElementAffinityScore(parentLocation, childLocation)
}

// EvaluatePersistentCacheParents sort persistent cache parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) EvaluatePersistentCacheParents(parents []*persistentcache.Peer, child *persistentcache.Peer, totalPieceCount uint32) []*persistentcache.Peer {
	sort.Slice(
		parents,
		func(i, j int) bool {
			return e.evaluatePersistentCacheParents(parents[i], child, totalPieceCount) > e.evaluatePersistentCacheParents(parents[j], child, totalPieceCount)
		},
	)

	return parents
}

// evaluatePersistentCacheParents sort persistent cache parents by evaluating multiple feature scores using custom algorithm.
func (e *customEvaluator) evaluatePersistentCacheParents(parent *persistentcache.Peer, child *persistentcache.Peer, totalPieceCount uint32) float64 {
	parentLocation := parent.Host.Network.Location
	parentIDC := parent.Host.Network.IDC
	childLocation := child.Host.Network.Location
	childIDC := child.Host.Network.IDC

	return customFinishedPieceWeight*e.calculatePieceScore(parent.FinishedPieces.Count(), child.FinishedPieces.Count(), totalPieceCount) +
		customIDCAffinityWeight*e.calculateIDCAffinityScore(parentIDC, childIDC) +
		customLocationAffinityWeight*e.calculateMultiElementAffinityScore(parentLocation, childLocation)
}

// calculatePieceScore 0.0~unlimited larger and better.
func (e *customEvaluator) calculatePieceScore(parentFinishedPieceCount uint, childFinishedPieceCount uint, totalPieceCount uint32) float64 {
	// If the total piece is determined, normalize the number of
	// pieces downloaded by the parent node.
	if totalPieceCount > 0 {
		return float64(parentFinishedPieceCount) / float64(totalPieceCount)
	}

	// Use the difference between the parent node and the child node to
	// download the piece to roughly represent the piece score.
	return float64(parentFinishedPieceCount) - float64(childFinishedPieceCount)
}

// calculateParentHostUploadSuccessScore 0.0~unlimited larger and better.
func (e *customEvaluator) calculateParentHostUploadSuccessScore(parentUploadCount int64, parentUploadFailedCount int64) float64 {
	if parentUploadCount < parentUploadFailedCount {
		return minScore
	}

	// Host has not been scheduled, then it is scheduled first.
	if parentUploadCount == 0 && parentUploadFailedCount == 0 {
		return maxScore
	}

	return float64(parentUploadCount-parentUploadFailedCount) / float64(parentUploadCount)
}

// calculateFreeUploadScore 0.0~1.0 larger and better.
func (e *customEvaluator) calculateFreeUploadScore(host *standard.Host) float64 {
	ConcurrentUploadLimit := host.ConcurrentUploadLimit.Load()
	freeUploadCount := host.FreeUploadCount()
	if ConcurrentUploadLimit > 0 && freeUploadCount > 0 {
		return float64(freeUploadCount) / float64(ConcurrentUploadLimit)
	}

	return minScore
}

// calculateHostTypeScore 0.0~1.0 larger and better.
func (e *customEvaluator) calculateHostTypeScore(peer *standard.Peer) float64 {
	// When the task is downloaded for the first time,
	// peer will be scheduled to seed peer first,
	// otherwise it will be scheduled to dfdaemon first.
	if peer.Host.Type != types.HostTypeNormal {
		if peer.FSM.Is(standard.PeerStateReceivedNormal) ||
			peer.FSM.Is(standard.PeerStateRunning) {
			return maxScore
		}

		return minScore
	}

	return maxScore * 0.5
}

// calculateIDCAffinityScore 0.0~1.0 larger and better.
func (e *customEvaluator) calculateIDCAffinityScore(dst, src string) float64 {
	if dst == "" || src == "" {
		return minScore
	}

	if strings.EqualFold(dst, src) {
		return maxScore
	}

	return minScore
}

// calculateMultiElementAffinityScore 0.0~1.0 larger and better.
func (e *customEvaluator) calculateMultiElementAffinityScore(dst, src string) float64 {
	if dst == "" || src == "" {
		return minScore
	}

	if strings.EqualFold(dst, src) {
		return maxScore
	}

	// Calculate the number of multi-element matches divided by "|".
	var score, elementLen int
	dstElements := strings.Split(dst, types.AffinitySeparator)
	srcElements := strings.Split(src, types.AffinitySeparator)
	elementLen = min(len(dstElements), len(srcElements))

	// Maximum element length is 5.
	elementLen = min(elementLen, maxElementLen)

	for i := range elementLen {
		if !strings.EqualFold(dstElements[i], srcElements[i]) {
			break
		}

		score++
	}

	return float64(score) / float64(maxElementLen)
} 