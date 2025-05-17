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

import (
	"context"
	"strings"
	"sync"
	"time"
)

type CacheVal struct {
	v    string
	when time.Time
	ttl  time.Duration
}

type InMemCache struct {
	mu     *sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	o      map[string]CacheVal
}

func NewInMemCache(ctx context.Context) *InMemCache {
	childCtx, cancel := context.WithCancel(ctx)
	v := &InMemCache{
		mu:     new(sync.RWMutex),
		ctx:    childCtx,
		cancel: cancel,
		o:      make(map[string]CacheVal),
	}

	go v.worker()

	return v
}

func (inMemCache *InMemCache) worker() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-inMemCache.ctx.Done():
			return
		case <-t.C:
			func() {
				inMemCache.mu.Lock()
				defer inMemCache.mu.Unlock()
				for k, v := range inMemCache.o {
					if v.ttl != 0 && time.Since(v.when) > v.ttl {
						delete(inMemCache.o, k)
					}
				}
			}()
		}
	}
}

func (inMemCache *InMemCache) SetWithExpire(key string, value string, ttl time.Duration) {
	inMemCache.mu.Lock()
	defer inMemCache.mu.Unlock()
	inMemCache.o[key] = CacheVal{
		v:    value,
		when: time.Now().UTC(),
		ttl:  ttl,
	}
}

func (inMemCache *InMemCache) Set(key string, value string) {
	inMemCache.mu.Lock()
	defer inMemCache.mu.Unlock()

	inMemCache.o[key] = CacheVal{
		v:    value,
		when: time.Now().UTC(),
	}
}

func (inMemCache *InMemCache) Get(key string) (string, bool) {
	inMemCache.mu.RLock()
	defer inMemCache.mu.RUnlock()

	val, ok := inMemCache.o[key]
	if !ok {
		return "", false
	}
	return val.v, true
}

func (inMemCache *InMemCache) Close() {
	if inMemCache.cancel != nil {
		inMemCache.cancel()
		clear(inMemCache.o)
	}
}

func (inMemCache *InMemCache) KeysWithPrefix(prefix string) ([]string, error) {
	inMemCache.mu.RLock()
	defer inMemCache.mu.RUnlock()

	var keys []string
	for k := range inMemCache.o {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
