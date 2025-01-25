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
	bootstrapHandler "github.com/ksctl/ksctl/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/pkg/consts"
	providerHandler "github.com/ksctl/ksctl/pkg/provider/handler"
)

func (kc *Controller) Delete() error {
	defer kc.b.PanicHandler(kc.l)

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

	{
		/*
		  Note: This will remove infrastructure created by the ksctl agent and not ksctl cli
		  CAUTION: WIP
		*/
		_, err := providerHandler.NewController(
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

		transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
		if errState != nil {
			kc.l.Error("handled error", "catch", errState)
			return err
		}

		kbc, errBootstrapController := bootstrapHandler.NewController(
			kc.ctx,
			kc.l,
			kc.b,
			kc.s,
			consts.OperationGet,
			&transferableInfraState,
			kc.p,
		)
		if errBootstrapController != nil {
			kc.l.Error("handled error", "catch", errBootstrapController)
			return errBootstrapController
		}

		if err := kbc.InvokeDestroyProcedure(); err != nil {
			kc.l.Error("handled error", "catch", err)
			return err
		}
	}

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationDelete,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	if errDelete := kpc.DeleteHACluster(); errDelete != nil {
		kc.l.Error("handled error", "catch", errDelete)
		return errDelete
	}

	kc.l.Success(kc.ctx, "successfully deleted ha cluster")

	return nil
}
