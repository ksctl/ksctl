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

package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksctl/ksctl/v2/pkg/statefile"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlError "github.com/ksctl/ksctl/v2/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"
)

func (p *Provider) generateConfig(noWorker, noControl int, cni bool) ([]byte, error) {
	if noWorker >= 0 && noControl == 0 {
		return nil, ksctlError.WrapError(
			ksctlError.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "invalid config request control node cannot be 0"),
		)
	}
	var config string
	config += `---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
`
	for noControl > 0 {
		config += `- role: control-plane
`
		noControl--
	}

	for noWorker > 0 {
		config += `- role: worker
`
		noWorker--
	}

	config += fmt.Sprintf(`networking:
  disableDefaultCNI: %v
`, cni)

	config += `...`

	return []byte(config), nil
}

func (p *Provider) configOption(noOfNodes int, cni bool) (cluster.CreateOption, error) {

	if noOfNodes < 1 {
		return nil, ksctlError.WrapError(
			ksctlError.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "invalid config request control node cannot be 0"),
		)
	}
	if noOfNodes == 1 {
		var config string
		config += fmt.Sprintf(`---
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
networking:
  disableDefaultCNI: %v
...`, cni)
		return cluster.CreateWithRawConfig([]byte(config)), nil
	}
	//control := noOfNodes / 2 // derive the math
	control := 1
	worker := noOfNodes - control
	raw, err := p.generateConfig(worker, control, cni)
	if err != nil {
		return nil, err
	}

	return cluster.CreateWithRawConfig(raw), nil
}

func (p *Provider) isPresent() error {
	return p.store.AlreadyCreated(consts.CloudLocal, "LOCAL", p.ClusterName, consts.ClusterTypeMang)
}

func (p *Provider) createNecessaryConfigs(storeDir string) (string, error) {
	_path := filepath.Join(storeDir, "kubeconfig")

	_, err := os.Create(_path)
	if err != nil {
		return "", ksctlError.WrapError(
			ksctlError.ErrInternal,
			p.l.NewError(p.ctx, "failed to create file to store kubeconfig", "Reason", err),
		)
	}
	return _path, nil
}

func (p *Provider) loadStateHelper() error {
	raw, err := p.store.Read()
	if err != nil {
		return err
	}
	*p.state = func(x *statefile.StorageDocument) statefile.StorageDocument {
		return *x
	}(raw)
	return nil
}
