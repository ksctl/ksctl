package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestArgocdComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall := setArgocdComponentOverridings(nil)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "latest",
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithNoUITrue(t *testing.T) {
	params := metadata.ComponentOverrides{
		"noUI": true,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithNoUIFalse(t *testing.T) {
	params := metadata.ComponentOverrides{
		"noUI": false,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/core-install.yaml"}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/stable/operator-manual/core/")
}

func TestArgocdComponentOverridingsWithNamespaceInstallTrue(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": true,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t,
		[]string{
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/appproject-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/namespace-install.yaml",
		}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability")
}

func TestArgocdComponentOverridingsWithNamespaceInstallFalse(t *testing.T) {
	params := metadata.ComponentOverrides{
		"namespaceInstall": false,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithVersionAndNoUI(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
		"noUI":    true,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithVersionAndNamespaceInstall(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":          "v1.0.0",
		"namespaceInstall": true,
	}
	version, url, postInstall := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t,
		[]string{
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/appproject-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/namespace-install.yaml",
		}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/v1.0.0/operator-manual/installation/#non-high-availability")
}
