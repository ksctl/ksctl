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

package cache

import "time"

type Cache interface {
	SetWithExpire(key string, value string, ttl time.Duration)
	Set(key string, value string)

	// Keys supports only glob-style patterns like *, ? and [chars].
	KeysWithPrefix(string) ([]string, error)

	Get(key string) (string, bool)
	Close()
}
