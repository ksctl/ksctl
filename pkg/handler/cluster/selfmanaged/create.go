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

package selfmanaged

import (
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/validation"

	bootstrapHandler "github.com/ksctl/ksctl/pkg/bootstrap/handler"

	providerHandler "github.com/ksctl/ksctl/pkg/providers/handler"
)

func (kc *Controller) Create() error {

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		consts.ClusterTypeHa,
	); err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	if !validation.ValidCNIPlugin(consts.KsctlValidCNIPlugin(kc.p.Metadata.CNIPlugin.StackName)) {
		err := kc.l.NewError(kc.ctx, "invalid CNI plugin")
		kc.l.Error("handled error", "catch", err)
		return err
	}

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationCreate,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	transferableInfraState, errProvisioning := kpc.CreateHACluster()
	if errProvisioning != nil {
		kc.l.Error("handled error", "catch", errProvisioning)
		return errProvisioning
	}

	kbc, errBootstrapController := bootstrapHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationCreate,
		transferableInfraState,
		kc.p,
	)
	if errBootstrapController != nil {
		kc.l.Error("handled error", "catch", errBootstrapController)
		return errBootstrapController
	}

	externalCNI, errConfiguration := kbc.ConfigureCluster()
	if errConfiguration != nil {
		kc.l.Error("handled error", "catch", errConfiguration)
		return errConfiguration
	}

	if err := kbc.InstallAdditionalTools(externalCNI); err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	kc.l.Success(kc.ctx, "successfully created ha cluster")

	return nil
}
