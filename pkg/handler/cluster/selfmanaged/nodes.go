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
	"strings"

	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	providerHandler "github.com/ksctl/ksctl/v2/pkg/provider/handler"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/validation"
)

func (kc *Controller) AddWorkerNodes() (errC error) {
	defer func() {
		v := kc.b.PanicHandler(kc.l)
		if v != nil {
			errC = errors.Join(errC, v)
			if kc.s.PlatformSpec.State != statefile.ConfiguringFailed {
				kc.s.PlatformSpec.State = statefile.ConfiguringFailed
				if err := kc.p.Storage.Write(kc.s); err != nil {
					errC = errors.Join(errC, err)
					kc.l.Error("Failed to write state after error", "error", err)
				}
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

	if state, err := kc.p.Storage.Read(); err != nil {
		if !ksctlErrors.IsNoMatchingRecordsFound(err) {
			return err
		}

		kc.l.Debug(kc.ctx, "No previous state found, creating a new one")

		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			kc.l.NewError(
				kc.ctx, "No previous state found",
			),
		)
	} else {
		kc.l.Debug(kc.ctx, "Found previous state, using it")
		if errOp := state.PlatformSpec.State.IsControllerOperationAllowed(consts.OperationScale); errOp != nil {
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

	defer func() {
		if errC != nil {
			kc.s.PlatformSpec.State = statefile.ConfiguringFailed
			if err := kc.p.Storage.Write(kc.s); err != nil {
				errC = errors.Join(errC, err)
				kc.l.Error("Failed to write state after error", "error", err)
			}
		} else {
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
		consts.OperationScale,
		kc.p,
	)
	if err != nil {
		return err
	}

	transferableInfraState, idxWPNotConfigured, errProvisioningWorker := kpc.AddWorkerNodes()
	if errProvisioningWorker != nil {
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
		return errBootstrapController
	}

	if err := kbc.JoinMoreWorkerPlanes(idxWPNotConfigured, kc.p.Metadata.NoWP); err != nil {
		return err
	}

	return nil
}

func (kc *Controller) DeleteWorkerNodes() (errC error) {
	defer func() {
		v := kc.b.PanicHandler(kc.l)
		if v != nil {
			errC = errors.Join(errC, v)
			if kc.s.PlatformSpec.State != statefile.ConfiguringFailed {
				kc.s.PlatformSpec.State = statefile.ConfiguringFailed
				if err := kc.p.Storage.Write(kc.s); err != nil {
					errC = errors.Join(errC, err)
					kc.l.Error("Failed to write state after error", "error", err)
				}
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

	if state, err := kc.p.Storage.Read(); err != nil {
		if !ksctlErrors.IsNoMatchingRecordsFound(err) {
			return err
		}

		kc.l.Debug(kc.ctx, "No previous state found, creating a new one")

		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			kc.l.NewError(
				kc.ctx, "No previous state found",
			),
		)
	} else {
		kc.l.Debug(kc.ctx, "Found previous state, using it")
		if errOp := state.PlatformSpec.State.IsControllerOperationAllowed(consts.OperationScale); errOp != nil {
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
	defer func() {
		if errC != nil {
			kc.s.PlatformSpec.State = statefile.ConfiguringFailed
			if err := kc.p.Storage.Write(kc.s); err != nil {
				errC = errors.Join(errC, err)
				kc.l.Error("Failed to write state after error", "error", err)
			}
		} else {
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
		consts.OperationScale,
		kc.p,
	)
	if err != nil {
		return err
	}

	transferableInfraState, hostnames, errDelWP := kpc.DelWorkerNodes()
	if errDelWP != nil {
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
			return err
		}

		if err := kbc.DelWorkerPlanes(kc.s.ClusterKubeConfig, hostnames); err != nil {
			return err
		}
	}

	return nil
}
