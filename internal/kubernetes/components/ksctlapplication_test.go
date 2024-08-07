package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestKsctlApplicationComponentOverridingsWithUriOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"uri": "https://example.com/custom/deploy.yml",
	}
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "main", version)
	assert.Equal(t, "https://example.com/custom/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}

func TestKsctlApplicationComponentOverridingsWithVersionAndUri(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
		"uri":     "https://example.com/custom/deploy.yml",
	}
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://example.com/custom/deploy.yml", url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}

func TestKsctlApplicationComponentOverridingsWithInvalidParams(t *testing.T) {

	loc := filepath.Join(os.TempDir(), "deploy.yml")

	params := metadata.ComponentOverrides{
		"version": "f14cd9094b2160c40ef8734e90141df81c22999e",
		"uri":     "file:::" + loc,
	}

	data := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
data:
  key: value
---
apiVersion: v1
kind: Pod
metadata:
  name: demo
spec:
  containers:
  - name: demo
    image: nginx
`)
	if err := os.WriteFile(loc, data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(loc); err != nil {
			t.Fatal(err)
		}
	})

	version, url, postInstall := setKsctlApplicationComponentOverridings(params)
	assert.Equal(t, "f14cd9094b2160c40ef8734e90141df81c22999e", version)
	assert.Equal(t, "file:::"+loc, url)
	assert.Equal(t, "As the controller and the crd are installed just need to apply application to be installed", postInstall)
}
