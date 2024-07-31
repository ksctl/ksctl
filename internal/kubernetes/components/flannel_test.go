package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlannelComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall := setFlannelComponentOverridings(nil)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, url, postInstall := setFlannelComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall := setFlannelComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/v1.0.0/download/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}
