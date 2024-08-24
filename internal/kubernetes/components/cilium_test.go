package components

import (
	"testing"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
)

func TestCiliumComponentOverridingsWithNilParams(t *testing.T) {
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(nil)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.Nil(t, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.Nil(t, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Nil(t, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithCiliumChartOverridingsOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"ciliumChartOverridings": map[string]any{"key": "value"},
	}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.NotNil(t, ciliumChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithVersionAndCiliumChartOverridings(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":                "v1.0.0",
		"ciliumChartOverridings": map[string]any{"key": "value"},
	}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.NotNil(t, ciliumChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, ciliumChartOverridings)
}
