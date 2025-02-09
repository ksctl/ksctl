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

package utilities

func DeduplicateStringsAlreadySorted[T string | int](input []T) []T {
	if len(input) < 2 {
		return input
	}
	j := 0
	for i := 1; i < len(input); i++ {
		if input[j] == input[i] {
			continue
		}
		j++
		input[j] = input[i]
	}
	return input[:j+1]
}
