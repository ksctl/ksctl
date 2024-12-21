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
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/providers/aws"
	"github.com/ksctl/ksctl/pkg/providers/azure"
	"github.com/ksctl/ksctl/pkg/providers/civo"
)

func (kc *Controller) Switch() (*string, error) {
	defer kc.b.PanicCatcher(kc.l)

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
		client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, localPkg.ProvideClient)

	}

	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	if err := client.Cloud.IsPresent(client.Storage); err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig, err := client.Cloud.GetKubeconfig(client.Storage)
	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	} else {
		if kubeconfig == nil {
			err = ksctlErrors.ErrKubeconfigOperations.Wrap(
				manager.log.NewError(
					controllerCtx, "Problem in kubeconfig get"),
			)

			log.Error("Kubeconfig we got is nil")
			return nil, err
		}
	}

	path, err := helpers.WriteKubeConfig(controllerCtx, *kubeconfig)
	log.Debug(controllerCtx, "data", "kubeconfigPath", path)

	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	printKubeConfig(manager.log, path)

	return kubeconfig, nil
}
