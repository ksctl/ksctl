package kubernetes

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/logger"
	"gotest.tools/v3/assert"
	"os"
	"testing"
)

// Here we are going to test the helper functions

func TestMain(m *testing.M) {

	log = logger.NewDefaultLogger(-1, os.Stdout)
	initApps()
	m.Run()
}

func TestConsts(t *testing.T) {
	assert.Equal(t, string(InstallHelm), "helm")
	assert.Equal(t, string(InstallKubectl), "kubectl")

	assert.Equal(t, string(Cni), "cni")
	assert.Equal(t, string(App), "app")
}

func TestGetApp(t *testing.T) {
	testCase := map[string]struct {
		appName         string
		version         string
		expectedToExist bool
	}{
		"argocd@v1.1.1": {
			appName:         "argocd",
			version:         "v1.1.1",
			expectedToExist: true,
		},
		"cilium@latest": {
			appName:         "cilium",
			version:         "latest",
			expectedToExist: true,
		},
		"abcd": {
			appName:         "",
			version:         "",
			expectedToExist: false,
		},
	}

	for app, expect := range testCase {
		_app, _ := helpers.ToApplicationTempl([]string{app})
		got, err := GetApps(_app[0].Name, _app[0].Version)
		v := err == nil // it will be true if there is no error
		assert.Equal(t, v, expect.expectedToExist)
		assert.Equal(t, got.Name, expect.appName)
		assert.Equal(t, got.Version, expect.version)
	}
}
