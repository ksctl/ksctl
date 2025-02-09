// Copyright 2025 Ksctl Authors
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

package utilities_test

import (
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/utilities"
	"gotest.tools/v3/assert"
)

func TestDeduplicateStringsAlreadySorted(t *testing.T) {
	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{"a", "a", "b", "b", "c", "c", "d", "d", "e", "e"}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{"a", "b", "c", "d", "e"})
	})

	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{"a", "b", "c", "d", "e"}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{"a", "b", "c", "d", "e"})
	})

	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{"a", "a", "b"}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{"a", "b"})
	})

	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{"a", "a"}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{"a"})
	})

	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{"a"}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{"a"})
	})

	t.Run("DeduplicateStringsAlreadySorted", func(t *testing.T) {
		input := []string{}
		output := utilities.DeduplicateStringsAlreadySorted(input)
		assert.DeepEqual(t, output, []string{})
	})
}
