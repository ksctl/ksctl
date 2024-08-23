package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestFlannelComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall, err := setFlannelComponentOverridings(nil)
	assert.Nil(t, err)
	assert.Equal(t, "v0.25.5", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/v0.25.5/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v0.25.5", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/v0.25.5/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/v1.0.0/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}
