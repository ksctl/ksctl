package kubernetes

import (
	"fmt"
	"os"
	"testing"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/logger"
	"gotest.tools/v3/assert"
)

// Here we are going to test the helper functions

func TestMain(m *testing.M) {

	log = logger.NewStructuredLogger(-1, os.Stdout)
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

func TestPresentOrNot(t *testing.T) {
	dummyState := new(storageTypes.StorageDocument)
	dummyState.Addons.Cni = storageTypes.Application{Name: "cilium", Version: "latest"}

	dummyState.Addons.Apps = []storageTypes.Application{
		{
			Name:    "dummy1",
			Version: "",
		},
	}
	testCases := []struct {
		App               storageTypes.Application
		TypeOfApp         EnumApplication
		ExpectedIdx       int
		ExpectedIsPresent bool
	}{
		{
			App: storageTypes.Application{
				Name:    "dummy1",
				Version: "latest",
			},
			TypeOfApp:         App,
			ExpectedIsPresent: true,
			ExpectedIdx:       0,
		},
		{
			App: storageTypes.Application{
				Name:    "cilium",
				Version: "latest",
			},
			TypeOfApp:         Cni,
			ExpectedIsPresent: true,
			ExpectedIdx:       -1,
		},
		{
			App: storageTypes.Application{
				Name:    "abcd",
				Version: "latest",
			},
			TypeOfApp:         App,
			ExpectedIsPresent: false,
			ExpectedIdx:       -1,
		},
	}

	for _, testCase := range testCases {
		gotIdx, gotUpdatable := PresentOrNot(
			testCase.App,
			testCase.TypeOfApp,
			dummyState,
		)

		assert.Check(t, gotUpdatable == testCase.ExpectedIsPresent,
			fmt.Sprintf("App: %v, got: %v, expect: %v\n",
				testCase.App, gotUpdatable, testCase.ExpectedIsPresent))
		assert.Check(t, gotIdx == testCase.ExpectedIdx,
			fmt.Sprintf("App: %v, got: %v, expect: %v\n",
				testCase.App, gotIdx, testCase.ExpectedIdx))
	}
}
