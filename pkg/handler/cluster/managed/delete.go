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
	providerHandler "github.com/ksctl/ksctl/v2/pkg/provider/handler"
)

func (kc *Controller) Delete() (errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
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
		consts.OperationDelete,
		kc.p,
	)
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	errKpc := kpc.DeleteManagedCluster()
	if errKpc != nil {
		kc.l.Error("handled error", "catch", errKpc)
		return errKpc
	}

	kc.l.Success(kc.ctx, "successfully deleted managed cluster")
	return nil
}
