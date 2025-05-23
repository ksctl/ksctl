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

package meta

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/provider/azure"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

type AzureMeta struct {
	ctx            context.Context
	l              logger.Logger
	cred           azcore.TokenCredential
	subscriptionId string
}

func (m *AzureMeta) Connect(ksctlConfig context.Context) error {
	v, ok := config.IsContextPresent(ksctlConfig, consts.KsctlAzureCredentials)
	if !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			m.l.NewError(m.ctx, "missing azure credentials"),
		)
	}
	extractedCreds := statefile.CredentialsAzure{}
	if err := json.Unmarshal([]byte(v), &extractedCreds); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			m.l.NewError(m.ctx, "failed to get azure credentials", "reason", err),
		)
	}

	if err := azure.SetRequiredENV_VAR(
		extractedCreds.SubscriptionID,
		extractedCreds.TenantID,
		extractedCreds.ClientID,
		extractedCreds.ClientSecret,
	); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed to set required env vars", "reason", err),
		)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "defaultAzureCredential", "Reason", err),
		)
	}

	m.cred = cred
	m.subscriptionId = extractedCreds.SubscriptionID
	return nil
}

func NewAzureMeta(ctx context.Context, l logger.Logger) (*AzureMeta, error) {
	ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "azure-metadata")

	return &AzureMeta{
		ctx: ctx,
		l:   l,
	}, nil
}

func (m *AzureMeta) GetAvailableRegions() ([]provider.RegionOutput, error) {
	clientFactory, err := armsubscriptions.NewClientFactory(m.cred, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed in azure client", "Reason", err),
		)
	}

	pager := clientFactory.NewClient().NewListLocationsPager(
		m.subscriptionId,
		&armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil},
	)

	var regions []provider.RegionOutput
	ctx := context.Background()

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				m.l.NewError(m.ctx, "failed to advance page", "Reason", err),
			)
		}
		for _, v := range page.Value {
			regions = append(regions, provider.RegionOutput{
				Sku:  *v.Name,
				Name: *v.DisplayName,
			})
		}
	}
	return regions, nil
}

func (m *AzureMeta) GetPriceInstanceType(regionSku string, instanceType string) (*provider.InstanceRegionOutput, error) {
	priceVM, err := m.priceVM(regionSku, instanceType)
	if err != nil {
		return nil, err
	}

	priceDisks, err := m.priceDisks(regionSku, WithDefaultManagedDisk())
	if err != nil {
		return nil, err
	}

	if len(priceDisks) == 0 {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed to get disk price"),
		)
	}

	priceVM.Disk = priceDisks[0]

	return priceVM, nil
}

func (m *AzureMeta) GetAvailableInstanceTypes(regionSku string, clusterType consts.KsctlClusterType) ([]provider.InstanceRegionOutput, error) {
	vms, err := m.listOfVms(regionSku)
	if err != nil {
		return nil, err
	}

	priceVMs, err := m.priceVMs(regionSku, WithDefaultInstanceType())
	if err != nil {
		return nil, err
	}

	vDisk := provider.StorageBlockRegionOutput{}

	if clusterType == consts.ClusterTypeSelfMang {
		disks, err := m.listOfDisksStandardLRS_ESeries(regionSku)
		if err != nil {
			return nil, err
		}

		priceDisks, err := m.priceDisks(regionSku, WithDefaultManagedDisk())
		if err != nil {
			return nil, err
		}

		for i := range disks {
			disks[i].Price = priceDisks[0].Price
			vDisk = disks[i]
		}
	}

	var outVMs []provider.InstanceRegionOutput

	for i := range vms {
		if price, ok := priceVMs[vms[i].Sku]; ok {
			vms[i].Price = price.Price
			vms[i].Disk = vDisk
			outVMs = append(outVMs, vms[i])
		}
	}

	return outVMs, nil
}

func (m *AzureMeta) GetAvailableManagedK8sManagementOfferings(regionSku string, _ *string) ([]provider.ManagedClusterOutput, error) {
	return m.priceAksManagement(regionSku, WithDefaultAks())
}

func (m *AzureMeta) GetAvailableManagedK8sVersions(regionSku string) ([]string, error) {
	clientFactory, err := armcontainerservice.NewClientFactory(m.subscriptionId, m.cred, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed in azure client", "Reason", err),
		)
	}

	res, err := clientFactory.NewManagedClustersClient().
		ListKubernetesVersions(m.ctx, regionSku, nil)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			m.l.NewError(m.ctx, "failed to get managed kubernetes versions", "Reason", err),
		)
	}

	var versions []string
	for _, v := range res.Values {
		if v.IsPreview != nil && *v.IsPreview {
			continue
		}
		if v.Version == nil {
			continue
		}
		versions = append(versions, *v.Version)
	}

	m.l.Debug(m.ctx, "Managed K8s versions", "AksVersions", versions)

	return versions, nil
}

func (m *AzureMeta) GetAvailableManagedCNIPlugins(regionSku string) (addons.ClusterAddons, string, error) {
	v, d, err := m.listManagedOfferingsCNIPlugins(regionSku)
	if err != nil {
		return nil, "", err
	}

	m.l.Debug(m.ctx, "Managed CNI plugins", "CNIPlugins", v)
	return v, d, nil
}
