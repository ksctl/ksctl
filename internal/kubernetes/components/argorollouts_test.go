package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestArgorolloutsComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall, err := setArgorolloutsComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/download/v1.0.0/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}

func TestArgorolloutsComponentOverridingsWithNamespaceInstallTrueOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": true,
	}
	version, url, postInstall, err := setArgorolloutsComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.7.2", version)
	assert.Equal(t, "https://raw.githubusercontent.com/argoproj/argo-rollouts/v1.7.2/manifests/namespace-install.yaml", url)
	assert.Contains(t, postInstall, "https://argo-rollouts.readthedocs.io/en/v1.7.2/installation/#controller-installation")
}

func TestArgorolloutsComponentOverridingsWithNamespaceInstallFalseOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": false,
	}
	version, url, postInstall, err := setArgorolloutsComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.7.2", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/download/v1.7.2/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}

func TestArgorolloutsComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, url, postInstall, err := setArgorolloutsComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.7.2", version)
	assert.Equal(t, "https://github.com/argoproj/argo-rollouts/releases/download/v1.7.2/install.yaml", url)
	assert.Contains(t, postInstall, "Commands to execute to access Argo-Rollouts")
}
