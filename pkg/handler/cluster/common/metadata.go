// Copyright 2025 Ksctl Authors
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
	"errors"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	azureMeta "github.com/ksctl/ksctl/v2/pkg/provider/azure/meta"
)

func (kc *Controller) SyncMetadata() (_ []provider.RegionOutput, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	// clusterType := consts.ClusterTypeMang
	// if kc.b.IsSelfManaged(kc.p) {
	// 	clusterType = consts.ClusterTypeSelfMang
	// }

	var metadataFromProvisioner provider.ProvisionMetadata

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		metadataFromProvisioner, err = azureMeta.NewAzureMeta(kc.ctx, kc.l)

		// case consts.CloudAws:
		// metadataFromProvisioner, err = aws.NewAwsMeta(kc.ctx, kc.l)
	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	regions, err := metadataFromProvisioner.GetAvailableRegions()
	if err != nil {
		kc.l.Error("handled error", "catch", err)
	}

	return regions, nil
}
