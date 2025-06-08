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
	"errors"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/validation"

	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"

	providerHandler "github.com/ksctl/ksctl/v2/pkg/provider/handler"
)

func (kc *Controller) Create() (errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		consts.ClusterTypeSelfMang,
	); err != nil {
		return err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			if errC != nil {
				errC = errors.Join(errC, err)
			} else {
				errC = err
			}
		}
	}()

	if state, err := kc.p.Storage.Read(); err != nil {
		if !ksctlErrors.IsNoMatchingRecordsFound(err) {
			return err
		}

		kc.l.Debug(kc.ctx, "No previous state found, creating a new one")

		if errOp := statefile.Fresh.IsControllerOperationAllowed(consts.OperationCreate); errOp != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidUserInput,
				errOp,
			)
		}
	} else {
		kc.l.Debug(kc.ctx, "Found previous state, using it")
		if errOp := state.PlatformSpec.State.IsControllerOperationAllowed(consts.OperationCreate); errOp != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidUserInput,
				errOp,
			)
		}
	}

	if !validation.ValidateDistro(kc.p.Metadata.K8sDistro) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidBootstrapProvider,
			kc.l.NewError(
				kc.ctx, "Problem in validation", "bootstrap", kc.p.Metadata.K8sDistro,
			),
		)
	}

	if err := validation.IsValidKsctlClusterAddons(kc.ctx, kc.l, kc.p.Metadata.Addons); err != nil {
		return err
	}

	defer func() {
		if errC != nil {
			// failed in cluster creation
			kc.s.PlatformSpec.State = statefile.CreationFailed
			if err := kc.p.Storage.Write(kc.s); err != nil {
				errC = errors.Join(errC, err)
				kc.l.Error("Failed to write state after error", "error", err)
			}
		} else {
			// successful cluster creation
			kc.s.PlatformSpec.State = statefile.Running
			if err := kc.p.Storage.Write(kc.s); err != nil {
				errC = errors.Join(errC, err)
				kc.l.Error("Failed to write state after success", "error", err)
			}
		}
	}()

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationCreate,
		kc.p,
	)
	if err != nil {
		return err
	}

	transferableInfraState, errProvisioning := kpc.CreateHACluster()
	if errProvisioning != nil {
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
		return errBootstrapController
	}

	externalCNI, errConfiguration := kbc.ConfigureCluster()
	if errConfiguration != nil {
		return errConfiguration
	}

	if err := kbc.InstallAdditionalTools(externalCNI); err != nil {
		return err
	}

	return nil
}
