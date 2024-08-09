package utilities

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLatestRepoReleaseReturnsTagName(t *testing.T) {
	_, err := GetLatestRepoRelease("ksctl", "ksctl")
	assert.NoError(t, err)
}

func TestGetLatestRepoReleaseHandlesNon200Status(t *testing.T) {
	_, err := GetLatestRepoRelease("org", "repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to get the latest version from the Github API")
}
