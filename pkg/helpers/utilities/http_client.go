package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetLatestRepoRelease(org, repo string) (string, error) {
	resp, err := HandleHTTPCall("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", org, repo), 5, nil, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to get the latest version from the Github API. Status code: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed deserialize response body: %v", err)
	}

	return release.TagName, nil
}

func GetAllRepoReleases(org, repo string) ([]string, error) {
	resp, err := HandleHTTPCall("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", org, repo), 5, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get the latest version from the Github API. Status code: %d", resp.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed deserialize response body: %v", err)
	}

	var tags []string
	for _, r := range releases {
		tags = append(tags, r.TagName)
	}

	return tags, nil
}

func HandleHTTPCall(method string, url string, timeout time.Duration,
	body io.Reader, headers map[string]string,
) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create newRequest, method: %s, url: %s, Reason: %v", method, url, err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	v, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed in client http request, method: %s, url: %s, Reason: %v", method, url, err)
	}
	return v, nil
}
