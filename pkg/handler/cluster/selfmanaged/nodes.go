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
	"strings"

	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	providerHandler "github.com/ksctl/ksctl/v2/pkg/provider/handler"
)

func (kc *Controller) AddWorkerNodes() error {
	defer kc.b.PanicHandler(kc.l)

	if !kc.b.IsSelfManaged(kc.p) {
		err := kc.l.NewError(kc.ctx, "this feature is only for selfmanaged clusters")
		kc.l.Error("handled error", "catch", err)
		return err
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		consts.ClusterTypeSelfMang,
	); err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationGet,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	transferableInfraState, idxWPNotConfigured, errProvisioningWorker := kpc.AddWorkerNodes()
	if errProvisioningWorker != nil {
		kc.l.Error("handled error", "catch", errProvisioningWorker)
		return errProvisioningWorker
	}

	kbc, errBootstrapController := bootstrapHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationGet,
		transferableInfraState,
		kc.p,
	)
	if errBootstrapController != nil {
		kc.l.Error("handled error", "catch", errBootstrapController)
		return errBootstrapController
	}

	if err := kbc.JoinMoreWorkerPlanes(idxWPNotConfigured, kc.p.Metadata.NoWP); err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	kc.l.Success(kc.ctx, "worker nodes added successfully")

	return nil
}

func (kc *Controller) DeleteWorkerNodes() error {
	defer kc.b.PanicHandler(kc.l)

	if !kc.b.IsSelfManaged(kc.p) {
		err := kc.l.NewError(kc.ctx, "this feature is only for selfmanaged clusters")
		kc.l.Error("handled error", "catch", err)
		return err
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		consts.ClusterTypeSelfMang,
	); err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationGet,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	transferableInfraState, hostnames, errDelWP := kpc.DelWorkerNodes()
	if errDelWP != nil {
		kc.l.Error("handled error", "catch", errDelWP)
		return errDelWP
	}

	kc.l.Debug(kc.ctx, "K8s nodes to be deleted", "hostnames", strings.Join(hostnames, ";"))

	fakeClient := false
	if _, ok := config.IsContextPresent(kc.ctx, consts.KsctlTestFlagKey); ok {
		fakeClient = true
	}

	if !fakeClient {
		kbc, err := bootstrapHandler.NewController(
			kc.ctx,
			kc.l,
			kc.b,
			kc.s,
			consts.OperationGet,
			transferableInfraState,
			kc.p,
		)
		if err != nil {
			kc.l.Error("handled error", "catch", err)
			return err
		}

		if err := kbc.DelWorkerPlanes(kc.s.ClusterKubeConfig, hostnames); err != nil {
			kc.l.Error("handled error", "catch", err)
			return err
		}
	}

	kc.l.Success(kc.ctx, "Successfully deleted workerNodes")

	return nil
}
