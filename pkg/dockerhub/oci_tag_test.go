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

package dockerhub_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/dockerhub"
)

type mockHttpResponse struct {
	io.Reader
}

func (m *mockHttpResponse) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func TestGetReleases(t *testing.T) {
	releases := `{"results":[
{
	"name":"v0.1.0",
	"last_updated":"2019-01-01T00:00:00Z"
},
{
	"name":"v1.0.0",
	"last_updated":"2020-01-01T00:00:00Z"
},
{
	"name":"v1.0.1-rc1",
	"last_updated":"2023-01-02T00:00:00Z"
}
]}`
	resp := &mockHttpResponse{Reader: bytes.NewBufferString(releases)}
	tags, err := dockerhub.ExtractTags(resp)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %v", len(tags))
	}
	if tags[0] != "v1.0.1-rc1" {
		t.Errorf("Expected v1.0.1-rc1, got %v", tags[0])
	}
	if tags[1] != "v1.0.0" {
		t.Errorf("Expected v1.0.0, got %v", tags[1])
	}
	if tags[2] != "v0.1.0" {
		t.Errorf("Expected v0.1.0, got %v", tags[2])
	}
}

func TestGetReleasesLatestPusbliedAtButOldRelease(t *testing.T) {
	releases := `{"results":[
{
	"name":"v1.12.13",
	"last_updated":"2024-08-08T17:00:07Z"
},
{
	"name":"v1.15.9",
	"last_updated":"2024-07-30T13:25:02Z"
},
{
	"name":"v1.15.20",
	"last_updated":"2024-07-30T13:25:02Z"
},
{
	"name":"v1.15.1",
	"last_updated":"2024-06-26T17:50:17Z"
}
]}`
	resp := &mockHttpResponse{Reader: bytes.NewBufferString(releases)}
	tags, err := dockerhub.ExtractTags(resp)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tags) != 4 {
		t.Errorf("Expected 3 tags, got %v", len(tags))
	}
	if tags[0] != "v1.15.20" {
		t.Errorf("Expected v1.15.20, got %v", tags[0])
	}
	if tags[1] != "v1.15.9" {
		t.Errorf("Expected v1.15.9, got %v", tags[1])
	}
	if tags[2] != "v1.15.1" {
		t.Errorf("Expected v1.15.1, got %v", tags[2])
	}
	if tags[3] != "v1.12.13" {
		t.Errorf("Expected v1.12.13, got %v", tags[3])
	}
}

func TestGetReleasesLatestWithout_v_asPrefix(t *testing.T) {
	releases := `{"results":[
{
	"name":"1.12.13",
	"last_updated":"2024-08-08T17:00:07Z"
},
{
	"name":"1.15.9",
	"last_updated":"2024-07-30T13:25:02Z"
},
{
	"name":"1.15.20",
	"last_updated":"2024-07-30T13:25:02Z"
},
{
	"name":"1.15.1",
	"last_updated":"2024-06-26T17:50:17Z"
}
]}`
	resp := &mockHttpResponse{Reader: bytes.NewBufferString(releases)}
	tags, err := dockerhub.ExtractTags(resp)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tags) != 4 {
		t.Errorf("Expected 3 tags, got %v", len(tags))
	}
	if tags[0] != "1.15.20" {
		t.Errorf("Expected 1.15.20, got %v", tags[0])
	}
	if tags[1] != "1.15.9" {
		t.Errorf("Expected 1.15.9, got %v", tags[1])
	}
	if tags[2] != "1.15.1" {
		t.Errorf("Expected 1.15.1, got %v", tags[2])
	}
	if tags[3] != "1.12.13" {
		t.Errorf("Expected 1.12.13, got %v", tags[3])
	}
}
