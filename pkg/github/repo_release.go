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

package github

import (
	"encoding/json"
	"fmt"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

func ExtractReleases(io io.Reader) ([]string, error) {
	type ReleaseInfo struct {
		TagName     string    `json:"tag_name"`
		Draft       bool      `json:"draft"`
		Prerelease  bool      `json:"prerelease"`
		PublishedAt time.Time `json:"published_at"`
	}

	var releases []ReleaseInfo

	if err := json.NewDecoder(io).Decode(&releases); err != nil {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to deserialize response body: %v", err,
		)
	}

	isRepoRespectSemver := true
	for i := range releases {
		if !semver.IsValid(releases[i].TagName) {
			isRepoRespectSemver = false
			releases[i].TagName = semver.Canonical("v" + releases[i].TagName)
		}
	}

	sort.Slice(releases, func(i, j int) bool {
		cmp := semver.Compare(releases[i].TagName, releases[j].TagName)
		if cmp != 0 {
			return cmp > 0
		}
		return releases[i].PublishedAt.After(releases[j].PublishedAt)
	})

	tags := make([]string, 0, len(releases))

	for _, r := range releases {
		if !r.Draft && !r.Prerelease {
			v := r.TagName
			if !isRepoRespectSemver {
				v = strings.TrimPrefix(v, "v")
			}
			tags = append(tags, v)
		}
	}

	if len(tags) == 0 {
		for _, r := range releases {
			if r.Prerelease && !r.Draft {
				tags = append(tags, r.TagName)
			}
		}
	}

	return tags, nil
}

func HttpGetAllStableGithubReleases(org, repo string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", org, repo))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Failed to get the latest version from the Github API. Status code: %d", resp.StatusCode,
		)
	}

	v, err := ExtractReleases(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Unable to get any releases. Not Even prerelease, most probabilty release are not there or in draft",
		)
	}
	return v, nil
}
