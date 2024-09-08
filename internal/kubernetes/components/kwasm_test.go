package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestSetKwasmOperatorComponentOverridings_DefaultValues(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, overridings, err := setKwasmOperatorComponentOverridings(params)

	assert.NoError(t, err)
	assert.Equal(t, "latest", version)
	assert.Nil(t, overridings)
}

func TestSetKwasmOperatorComponentOverridings_WithOverrides(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.2.3",
		"kwasmOperatorChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	version, overridings, err := setKwasmOperatorComponentOverridings(params)

	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", version)
	assert.NotNil(t, overridings)
}

func TestKwasmWasmedgeComponent(t *testing.T) {
	params := metadata.ComponentOverrides{}
	component, err := KwasmComponent(params)

	assert.NoError(t, err)
	assert.Equal(t, metadata.ComponentTypeKubectl, component.HandlerType)
	assert.NotNil(t, component.Kubectl)
	assert.Equal(t, "latest", component.Kubectl.Version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/ksctl/components/main/wasm/kwasm/runtimeclass.yml"}, component.Kubectl.Urls)
}

func TestKwasmWithSpinKube(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.2.3",
		kwasmOperatorChartOverridingsKey: map[string]any{
			"someKey": "someValue",
		},
	}

	if err := GetSpinKubeStackSpecificKwasmOverrides(params); err != nil {
		t.Fatal(err)
	}

	version, overridings, err := setKwasmOperatorComponentOverridings(params)

	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", version)
	assert.NotNil(t, overridings)
	assert.Contains(t, overridings, "kwasmOperator")
	assert.Contains(t, overridings["kwasmOperator"].(map[string]any), "installerImage")
	assert.Equal(t, "someValue", overridings["someKey"])
	assert.Equal(
		t,
		overridings["kwasmOperator"].(map[string]any)["installerImage"],
		"ghcr.io/spinkube/containerd-shim-spin/node-installer:v0.15.1")
}
