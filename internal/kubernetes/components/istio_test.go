package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsitoComponentOverridingsWithNilParams(t *testing.T) {
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(nil)
	assert.Equal(t, "latest", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithEmptyParams(t *testing.T) {
	params := metadata.ComponentOverrides{}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithHelmBaseChartOverridingsOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"helmBaseChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithHelmIstiodChartOverridingsOnly(t *testing.T) {
	params := metadata.ComponentOverrides{
		"helmIstiodChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionAndHelmBaseChartOverridings(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":                  "v1.0.0",
		"helmBaseChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionAndHelmIstiodChartOverridings(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":                    "v1.0.0",
		"helmIstiodChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithAllParams(t *testing.T) {
	params := metadata.ComponentOverrides{
		"version":                    "v1.0.0",
		"helmBaseChartOverridings":   map[string]any{"baseKey": "baseValue"},
		"helmIstiodChartOverridings": map[string]any{"istiodKey": "istiodValue"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings := setIsitoComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"baseKey": "baseValue"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"istiodKey": "istiodValue"}, helmIstiodChartOverridings)
}
