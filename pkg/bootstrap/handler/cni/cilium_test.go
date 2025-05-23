// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cni

import (
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/stretchr/testify/assert"
)

func TestCiliumComponentOverridingsWithNilParams(t *testing.T) {
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(nil)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.Equal(t, map[string]any{
		"hubble": map[string]any{
			"ui":    map[string]any{"enabled": true},
			"relay": map[string]any{"enabled": true},
			"metrics": map[string]any{"enabled": []string{
				"dns",
				"drop",
				"tcp",
				"flow",
				"port-distribution",
				"icmp",
				"httpV2:exemplars=true;labelsContext=source_ip,source_namespace,source_workload,destination_ip,destination_namespace,destination_workload,traffic_direction",
			}},
		},
		"l7Proxy":              true,
		"kubeProxyReplacement": true,
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"operator": map[string]any{
			"replicas": 3,
			"prometheus": map[string]any{
				"enabled": true,
			},
		},
		"prometheus": map[string]any{
			"enabled": true,
		},
	}, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithEmptyParams(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.Equal(t, map[string]any{
		"hubble": map[string]any{
			"ui":    map[string]any{"enabled": true},
			"relay": map[string]any{"enabled": true},
			"metrics": map[string]any{"enabled": []string{
				"dns",
				"drop",
				"tcp",
				"flow",
				"port-distribution",
				"icmp",
				"httpV2:exemplars=true;labelsContext=source_ip,source_namespace,source_workload,destination_ip,destination_namespace,destination_workload,traffic_direction",
			}},
		},
		"l7Proxy":              true,
		"kubeProxyReplacement": true,
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"operator": map[string]any{
			"replicas": 3,
			"prometheus": map[string]any{
				"enabled": true,
			},
		},
		"prometheus": map[string]any{
			"enabled": true,
		},
	}, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithVersionOnly(t *testing.T) {

	t.Run("with version and GuidedConfig", func(t *testing.T) {

		params := stack.ComponentOverrides{
			"version":      "v1.0.0",
			"guidedConfig": []string{"GG", "l7-proxy"},
		}
		version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
		assert.Nil(t, err)
		assert.Equal(t, "v1.0.0", version)
		assert.Equal(t, map[string]any{
			"l7Proxy": true,
		}, ciliumChartOverridings)
	})

	t.Run("guidedConfig Hubble, encryption, and prometheus", func(t *testing.T) {
		params := stack.ComponentOverrides{
			"version":      "v1.0.0",
			"guidedConfig": []string{"hubble", "encryption", "prometheus"},
		}
		version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
		assert.Nil(t, err)
		assert.Equal(t, "v1.0.0", version)
		assert.Equal(t, map[string]any{
			"hubble": map[string]any{
				"ui":    map[string]any{"enabled": true},
				"relay": map[string]any{"enabled": true},
				"metrics": map[string]any{"enabled": []string{
					"dns",
					"drop",
					"tcp",
					"flow",
					"port-distribution",
					"icmp",
					"httpV2:exemplars=true;labelsContext=source_ip,source_namespace,source_workload,destination_ip,destination_namespace,destination_workload,traffic_direction",
				}},
			},
			"encryption": map[string]any{
				"enabled": true,
				"type":    "wireguard",
			},
			"operator": map[string]any{
				"replicas": 3,
				"prometheus": map[string]any{
					"enabled": true,
				},
			},
			"prometheus": map[string]any{
				"enabled": true,
			},
		}, ciliumChartOverridings)
	})
}

func TestCiliumComponentOverridingsWithCiliumChartOverridingsOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"ciliumChartOverridings": map[string]any{"key": "value"},
	}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.16.1", version)
	assert.NotNil(t, ciliumChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, ciliumChartOverridings)
}

func TestCiliumComponentOverridingsWithVersionAndCiliumChartOverridings(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":                "v1.0.0",
		"ciliumChartOverridings": map[string]any{"key": "value"},
	}
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.NotNil(t, ciliumChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, ciliumChartOverridings)
}
