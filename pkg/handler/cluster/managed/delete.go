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

package managed

import (
	"errors"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	providerHandler "github.com/ksctl/ksctl/v2/pkg/provider/handler"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

func (kc *Controller) Delete() (errC error) {
	defer func() {
		v := kc.b.PanicHandler(kc.l)
		if v != nil {
			errC = errors.Join(errC, v)
			if kc.s.PlatformSpec.State != statefile.DeletionFailed {
				kc.s.PlatformSpec.State = statefile.DeletionFailed
				if err := kc.p.Storage.Write(kc.s); err != nil {
					errC = errors.Join(errC, err)
					kc.l.Error("Failed to write state after error", "error", err)
				}
			}
		}
	}()

	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		consts.ClusterTypeMang,
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
		if errOp := state.PlatformSpec.State.IsControllerOperationAllowed(consts.OperationDelete); errOp != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidUserInput,
				errOp,
			)
		}
	}

	defer func() {
		if errC != nil {
			// failed in cluster deletion failed
			kc.s.PlatformSpec.State = statefile.DeletionFailed
			if err := kc.p.Storage.Write(kc.s); err != nil {
				errC = errors.Join(errC, err)
				kc.l.Error("Failed to write state after error", "error", err)
			}
		}
	}()

	kpc, err := providerHandler.NewController(
		kc.ctx,
		kc.l,
		kc.b,
		kc.s,
		consts.OperationDelete,
		kc.p,
	)
	if err != nil {
		return err
	}

	errKpc := kpc.DeleteManagedCluster()
	if errKpc != nil {
		return errKpc
	}

	return nil
}
