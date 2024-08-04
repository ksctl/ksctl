package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKsctlApplicationComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall := setKsctlApplicationComponentOverridings(nil)
	assert.Equal(t, "main", version)
	assert.Equal(t, "https://raw.githubusercontent.com/ksctl/ksctl/main/ksctl-components/manifests/controllers/application/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}

func TestKsctlApplicationComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "main", version)
	assert.Equal(t, "https://raw.githubusercontent.com/ksctl/ksctl/main/ksctl-components/manifests/controllers/application/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}

func TestKsctlApplicationComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://raw.githubusercontent.com/ksctl/ksctl/v1.0.0/ksctl-components/manifests/controllers/application/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}

func TestKsctlApplicationComponentOverridingsWithVersionLatest(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "latest",
	}
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "main", version)
	assert.Equal(t, "https://raw.githubusercontent.com/ksctl/ksctl/main/ksctl-components/manifests/controllers/application/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}
