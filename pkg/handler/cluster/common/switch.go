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

package common

import (
	"github.com/ksctl/ksctl/cli"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/providers/aws"
	"github.com/ksctl/ksctl/pkg/providers/azure"
	"github.com/ksctl/ksctl/pkg/providers/civo"
	"github.com/ksctl/ksctl/pkg/providers/local"
)

func (kc *Controller) Switch() (*string, error) {
	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	clusterType := consts.ClusterTypeMang
	if kc.b.IsHA(kc.p) {
		clusterType = consts.ClusterTypeHa
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		clusterType); err != nil {

		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudCivo:
		kc.p.Cloud, err = civo.NewClient(kc.ctx, kc.p.Metadata, kc.l, kc.s, civo.ProvideClient)

	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.p.Metadata, kc.l, kc.s, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.p.Metadata, kc.l, kc.s, aws.ProvideClient)
		if err != nil {
			break
		}

		err = cloudController.InitCloud(client, stateDocument, consts.OperationGet)

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.p.Metadata, kc.l, kc.s, local.ProvideClient)

	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig, err := kc.p.Cloud.GetKubeconfig()
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if kubeconfig == nil {
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			kc.l.NewError(
				kc.ctx, "Problem in kubeconfig get"),
		)

		kc.l.Error("Kubeconfig we got is nil")
		return nil, err
	}

	path, err := cli.WriteKubeConfig(kc.ctx, *kubeconfig)
	kc.l.Debug(kc.ctx, "data", "kubeconfigPath", path)

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	printKubeConfig(kc.l, path)

	return kubeconfig, nil
}
