// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utilities

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDeepCopyMap(t *testing.T) {
	t.Run("copy from one to another", func(t *testing.T) {
		src := map[string]int{
			"one": 1,
			"two": 2,
		}
		dest := DeepCopyMap(src)
		assert.DeepEqual(t, dest, src)
	})

	t.Run("Copy from entireely from new preserving the older vals in process", func(t *testing.T) {

		t.Run("Depth one", func(t *testing.T) {
			src := map[string]any{
				"one":   1,
				"two":   2,
				"three": 3,
			}

			override := map[string]any{
				"four": 4,
			}

			CopySrcToDestPreservingDestVals(override, src)
			assert.DeepEqual(t, override, map[string]any{
				"one":   1,
				"two":   2,
				"three": 3,
				"four":  4,
			})
		})

		t.Run("Depth two", func(t *testing.T) {
			src := map[string]any{
				"one": 1,
				"two": 2,
				"three": map[string]any{
					"four": 4,
				},
			}

			override := map[string]any{
				"three": map[string]any{
					"five": 5,
				},
			}

			CopySrcToDestPreservingDestVals(override, src)
			assert.DeepEqual(t, override, map[string]any{
				"one": 1,
				"two": 2,
				"three": map[string]any{
					"four": 4,
					"five": 5,
				},
			})
		})

		t.Run("Depth three", func(t *testing.T) {

			src := map[string]any{
				"one": 1,
				"two": 2,
				"three": map[string]any{
					"four": []int{1, 2, 3},
				},
			}

			override := map[string]any{
				"three": map[string]any{
					"four": []int{3, 5, 6},
					"five": 5,
				},
			}

			CopySrcToDestPreservingDestVals(override, src)
			assert.DeepEqual(t, override, map[string]any{
				"one": 1,
				"two": 2,
				"three": map[string]any{
					"four": []int{3, 5, 6, 1, 2},
					"five": 5,
				},
			})
		})
	})

}

func TestDeepCopySlice(t *testing.T) {
	src := []int{1, 2, 3}
	dest := DeepCopySlice(src)
	assert.DeepEqual(t, dest, src)
}
