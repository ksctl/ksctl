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
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/providers"
	"github.com/ksctl/ksctl/pkg/providers/aws"
	"github.com/ksctl/ksctl/pkg/providers/azure"
	"github.com/ksctl/ksctl/pkg/providers/civo"
	"github.com/ksctl/ksctl/pkg/providers/local"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (kc *Controller) clusterDataHelper(operation logger.LogClusterDetail) ([]logger.ClusterDataForLogging, error) {
	if err := kc.b.ValidateMetadata(kc.p); err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	if operation == logger.LoggingInfoCluster {
		var err []error
		if len(kc.p.Metadata.ClusterName) == 0 {
			err = append(err,
				ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidResourceName,
					kc.l.NewError(kc.ctx, "cluster name is needed for cluster info details"),
				),
			)
		}
		if len(kc.p.Metadata.Region) == 0 {
			err = append(err,
				ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidCloudRegion,
					kc.l.NewError(kc.ctx, "region is needed for cluster info details"),
				),
			)
		}
		if len(err) != 0 {
			_err := ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidUserInput,
				kc.l.NewError(kc.ctx, "Failure", "reason", err),
			)

			kc.l.Error("Failure", "reason", _err)

			return nil, _err
		}
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	kc.l.Note(kc.ctx, "Filter", "cloudProvider", string(kc.p.Metadata.Provider))

	var (
		cloudMapper = map[consts.KsctlCloud]providers.Cloud{
			consts.CloudCivo:  nil,
			consts.CloudAzure: nil,
			consts.CloudAws:   nil,
			consts.CloudLocal: nil,
		}
	)

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudCivo:
		cloudMapper[consts.CloudCivo], err = civo.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, civo.ProvideClient)

	case consts.CloudAzure:
		cloudMapper[consts.CloudAzure], err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		cloudMapper[consts.CloudAws], err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, aws.ProvideClient)

	case consts.CloudLocal:
		cloudMapper[consts.CloudLocal], err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, local.ProvideClient)

	default:
		switch operation {
		case logger.LoggingGetClusters:
			if kc.p.Metadata.Provider != consts.CloudAll {
				err = ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidCloudProvider,
					kc.l.NewError(
						kc.ctx, "", "cloud", kc.p.Metadata.Provider,
					),
				)
			} else {
				cloudMapper[consts.CloudCivo], err = civo.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, civo.ProvideClient)
				if err != nil {
					kc.l.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudAzure], err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, azure.ProvideClient)
				if err != nil {
					kc.l.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudAws], err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, aws.ProvideClient)
				if err != nil {
					kc.l.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudLocal], err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, local.ProvideClient)
				if err != nil {
					kc.l.Error("handled error", "catch", err)
					return nil, err
				}
			}

		case logger.LoggingInfoCluster:
			err = ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidCloudProvider,
				kc.l.NewError(
					kc.ctx, "", "cloud", kc.p.Metadata.Provider,
				),
			)
		}

	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	var printerTable []logger.ClusterDataForLogging
	for _, v := range cloudMapper {
		if v == nil {
			continue
		}
		data, err := v.GetRAWClusterInfos()
		if err != nil {
			kc.l.Error("handled error", "catch", err)
			return nil, err
		}

		if operation == logger.LoggingInfoCluster {
			// as the info will not have all as cloud provider this loop will only run once
			for _, _data := range data {
				if _data.Name != kc.p.Metadata.ClusterName ||
					_data.Region != kc.p.Metadata.Region {
					continue
				} else {
					printerTable = append(printerTable, _data)
					return printerTable, nil
				}
			}
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				kc.l.NewError(kc.ctx, "No state is present",
					"name", kc.p.Metadata.ClusterName),
			)
		}

		printerTable = append(printerTable, data...)
	}
	return printerTable, err
}

func (kc *Controller) GetCluster() error {
	v, err := kc.clusterDataHelper(logger.LoggingGetClusters)
	if err != nil {
		return err
	}

	kc.l.Table(kc.ctx, logger.LoggingGetClusters, v)

	kc.l.Success(kc.ctx, "successfully get clusters")

	return nil
}

func (kc *Controller) InfoCluster() (*logger.ClusterDataForLogging, error) {
	v, err := kc.clusterDataHelper(logger.LoggingInfoCluster)
	if err != nil {
		return nil, err
	}

	kc.l.Table(kc.ctx, logger.LoggingInfoCluster, v)

	kc.l.Success(kc.ctx, "successfully cluster info")

	return utilities.Ptr[logger.ClusterDataForLogging](v[0]), nil
}
