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

package dockerhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"golang.org/x/mod/semver"
)

func ExtractTags(io io.Reader) ([]string, error) {
	type Result struct {
		Results []struct {
			TagName   string    `json:"name"`
			UpdatedAt time.Time `json:"last_updated"`
		} `json:"results"`
	}

	var releases Result

	if err := json.NewDecoder(io).Decode(&releases); err != nil {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to deserialize response body: %v", err,
		)
	}

	res := releases.Results
	isRepoRespectSemver := true
	for i := range res {
		if !semver.IsValid(res[i].TagName) {
			isRepoRespectSemver = false
			res[i].TagName = semver.Canonical("v" + res[i].TagName)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		cmp := semver.Compare(res[i].TagName, res[j].TagName)
		if cmp != 0 {
			return cmp > 0
		}
		return res[i].UpdatedAt.After(res[j].UpdatedAt)
	})

	tags := make([]string, 0, len(res))

	for _, r := range res {
		v := r.TagName
		if !isRepoRespectSemver {
			v = strings.TrimPrefix(v, "v")
		}
		tags = append(tags, v)
	}

	return tags, nil
}

func HttpGetAllTags(org, repo string) ([]string, error) {
	uri := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags?page_size=20", org, repo)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Failed to get the latest version from the DockerHub. Status code: %d", resp.StatusCode,
		)
	}

	v, err := ExtractTags(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Unable to get any releases",
		)
	}
	return v, nil
}
