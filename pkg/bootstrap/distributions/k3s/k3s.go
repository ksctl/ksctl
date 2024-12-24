// Copyright 2024 ksctl
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

package k3s

import (
	"context"
	"sync"

	"github.com/ksctl/ksctl/pkg/bootstrap/distributions"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/ssh"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"

	"github.com/ksctl/ksctl/pkg/consts"
)

type K3s struct {
	ctx   context.Context
	l     logger.Logger
	state *statefile.StorageDocument
	mu    *sync.Mutex
	store storage.Storage

	Cni string
}

func NewClient(
	parentCtx context.Context,
	parentLog logger.Logger,
	storage storage.Storage,
	state *statefile.StorageDocument,
) *K3s {
	p := &K3s{mu: &sync.Mutex{}}
	p.ctx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.K8sK3s))
	p.l = parentLog
	p.state = state
	p.store = storage

	return p
}

func (p *K3s) Setup(operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		p.state.K8sBootstrap.K3s = &statefile.StateConfigurationK3s{}
		p.state.BootstrapProvider = consts.K8sK3s
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}
	return nil
}

func scriptKUBECONFIG() ssh.ExecutionPipeline {
	collection := ssh.NewExecutionPipeline()
	collection.Append(ssh.Script{
		Name:           "k3s kubeconfig",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo cat /etc/rancher/k3s/k3s.yaml
`,
	})

	return collection
}

func (p *K3s) K8sVersion(ver string) distributions.KubernetesDistribution {
	if v, err := p.isValidK3sVersion(ver); err == nil {
		p.state.K8sBootstrap.K3s.K3sVersion = v
		p.l.Debug(p.ctx, "Printing", "k3s.K3sVer", v)
		return p
	} else {
		p.l.Error(err.Error())
		return nil
	}
}

func (p *K3s) CNI(cni string) (externalCNI bool) {
	p.l.Debug(p.ctx, "Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case consts.CNIFlannel, "":
		p.Cni = string(consts.CNIFlannel)
		return false

	default:
		// this tells us that CNI should be installed via the k8s client
		p.Cni = string(consts.CNINone)
		return true
	}
}
