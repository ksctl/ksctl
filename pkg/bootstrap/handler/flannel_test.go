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

package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlannelComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall, err := setFlannelComponentOverridings(nil)
	assert.Nil(t, err)
	assert.Equal(t, "v0.25.5", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/download/v0.25.5/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithEmptyParams(t *testing.T) {
	params := ComponentOverrides{}
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v0.25.5", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/download/v0.25.5/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}

func TestFlannelComponentOverridingsWithVersionOnly(t *testing.T) {
	params := ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "https://github.com/flannel-io/flannel/releases/download/v1.0.0/kube-flannel.yml", url)
	assert.Equal(t, "https://github.com/flannel-io/flannel", postInstall)
}
