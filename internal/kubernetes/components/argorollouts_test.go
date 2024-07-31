package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArgorolloutsComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall := setArgorolloutsComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/v1.0.0/download/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}

func TestArgorolloutsComponentOverridingsWithNamespaceInstallTrueOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": true,
	}
	version, url, postInstall := setArgorolloutsComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "https://raw.githubusercontent.com/argoproj/argo-cd/latest/manifests/namespace-install.yaml", url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/latest/operator-manual/installation/#non-high-availability")
}

func TestArgorolloutsComponentOverridingsWithNamespaceInstallFalseOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": false,
	}
	version, url, postInstall := setArgorolloutsComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}

func TestArgorolloutsComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, url, postInstall := setArgorolloutsComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}
