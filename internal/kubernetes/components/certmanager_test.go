package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestCertManagerComponentWithNilParams(t *testing.T) {
	params := metadata.ComponentOverrides(nil)
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "cert-manager", component.Helm.Charts[0].ReleaseName)
	assert.Equal(t, "cert-manager", component.Helm.Charts[0].Namespace)
	assert.Equal(t, "https://charts.jetstack.io", component.Helm.RepoUrl)
	assert.Equal(t, "jetstack", component.Helm.RepoName)

	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, "true", v)
		}
	}
}

func TestCertManagerComponentWithVersionOverride(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", component.Helm.Charts[0].Version)
}

func TestCertManagerComponentWithGatewayApiEnable(t *testing.T) {
	params := metadata.ComponentOverrides{
		"gatewayapiEnable": true,
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Contains(t, component.Helm.Charts[0].Args["extraArgs"], "--enable-gateway-api")
}

func TestCertManagerComponentWithCertManagerChartOverridings(t *testing.T) {
	params := metadata.ComponentOverrides{
		"certmanagerChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "someValue", component.Helm.Charts[0].Args["someKey"])

	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, "true", v)
		}
	}
}

func TestCertManagerComponentWithAllOverrides(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":          "v1.0.0",
		"gatewayapiEnable": true,
		"certmanagerChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", component.Helm.Charts[0].Version)
	assert.Contains(t, component.Helm.Charts[0].Args["extraArgs"], "--enable-gateway-api")
	assert.Equal(t, "someValue", component.Helm.Charts[0].Args["someKey"])
	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, "true", v)
		}
	}
}
