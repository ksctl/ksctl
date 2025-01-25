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

//go:build poller_real

package poller_test

import (
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/v2/pkg/poller"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleasePoller()
	m.Run()
}

func TestGithub(t *testing.T) {
	obj := poller.GetSharedPoller()

	r, err := obj.Get("flannel-io", "flannel")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	dump.Println(r)
}
