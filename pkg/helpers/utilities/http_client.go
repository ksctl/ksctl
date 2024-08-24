package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

func extractReleases(io io.Reader) ([]string, error) {
	type ReleaseInfo struct {
		TagName     string    `json:"tag_name"`
		Draft       bool      `json:"draft"`
		Prerelease  bool      `json:"prerelease"`
		PublishedAt time.Time `json:"published_at"`
	}

	var releases []ReleaseInfo

	if err := json.NewDecoder(io).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to deserialize response body: %w", err)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get the latest version from the Github API. Status code: %d", resp.StatusCode)
	}

	v, err := extractReleases(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to extract the releases: %w", err)
	}
	if len(v) == 0 {
		return nil, fmt.Errorf("Unable to get any releases. Not Even prerelease, most probabilty release are not there or in draft")
	}
	return v, nil
}
