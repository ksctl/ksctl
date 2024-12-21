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
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (kc *Controller) clusterDataHelper(operation consts.LogClusterDetail) ([]cloudControllerResource.AllClusterData, error) {

	client := kc.client
	log := kc.log
	defer panicCatcher(log)

	if err := kc.validationFields(client.Metadata); err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	if client.Metadata.Provider == consts.CloudLocal {
		client.Metadata.Region = "LOCAL"
	}

	if operation == consts.LoggingInfoCluster {
		var err []error
		if len(client.Metadata.ClusterName) == 0 {
			err = append(err,
				ksctlErrors.ErrInvalidResourceName.Wrap(
					log.NewError(controllerCtx, "clustername is needed for cluster info details"),
				),
			)
		}
		if len(client.Metadata.Region) == 0 {
			err = append(err,
				ksctlErrors.ErrInvalidCloudRegion.Wrap(
					log.NewError(controllerCtx, "region is needed for cluster info details"),
				),
			)
		}
		if len(err) != 0 {
			_err := ksctlErrors.ErrInvalidUserInput.Wrap(
				log.NewError(controllerCtx, "Failure", "reason", err),
			)

			log.Error("Failure", "reason", _err)

			return nil, _err
		}
	}

	defer func() {
		if err := client.Storage.Kill(); err != nil {
			log.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	log.Note(controllerCtx, "Filter", "cloudProvider", string(client.Metadata.Provider))

	var (
		cloudMapper = map[consts.KsctlCloud]types.CloudFactory{
			consts.CloudCivo:  nil,
			consts.CloudAzure: nil,
			consts.CloudAws:   nil,
			consts.CloudLocal: nil,
		}
	)

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)

	case consts.CloudAzure:
		cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)

	case consts.CloudAws:
		cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)

	case consts.CloudLocal:
		cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)

	default:
		switch operation {
		case consts.LoggingGetClusters:
			if client.Metadata.Provider != consts.CloudAll {
				err = ksctlErrors.ErrInvalidCloudProvider.Wrap(
					kc.log.NewError(
						controllerCtx, "", "cloud", client.Metadata.Provider,
					),
				)
			} else {
				cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)
				if err != nil {
					log.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)
				if err != nil {
					log.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)
				if err != nil {
					log.Error("handled error", "catch", err)
					return nil, err
				}
				cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)
				if err != nil {
					log.Error("handled error", "catch", err)
					return nil, err
				}
			}

		case consts.LoggingInfoCluster:
			err = ksctlErrors.ErrInvalidCloudProvider.Wrap(
				kc.log.NewError(
					controllerCtx, "", "cloud", client.Metadata.Provider,
				),
			)
		}

	}

	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	var printerTable []cloudControllerResource.AllClusterData
	for _, v := range cloudMapper {
		if v == nil {
			continue
		}
		data, err := v.GetRAWClusterInfos(client.Storage)
		if err != nil {
			log.Error("handled error", "catch", err)
			return nil, err
		}

		if operation == consts.LoggingInfoCluster {
			// as the info will not have all as cloud provider this loop will only run once
			for _, _data := range data {
				if _data.Name != kc.client.Metadata.ClusterName ||
					_data.Region != kc.client.Metadata.Region {
					continue
				} else {
					printerTable = append(printerTable, _data)
					return printerTable, nil
				}
			}
			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
				log.NewError(controllerCtx, "No state is present",
					"name", kc.client.Metadata.ClusterName),
			)
		}

		printerTable = append(printerTable, data...)
	}
	return printerTable, err
}

func (kc *Controller) GetCluster() error {
	v, err := kc.clusterDataHelper(
		consts.LoggingGetClusters,
	)
	if err != nil {
		return err
	}

	kc.log.Table(controllerCtx, consts.LoggingGetClusters, v)

	kc.log.Success(controllerCtx, "successfully get clusters")

	return nil
}

func (kc *Controller) InfoCluster() (*cloudControllerResource.AllClusterData, error) {
	v, err := kc.clusterDataHelper(
		consts.LoggingInfoCluster,
	)
	if err != nil {
		return nil, err
	}

	kc.log.Table(controllerCtx, consts.LoggingInfoCluster, v)

	kc.log.Success(controllerCtx, "successfully cluster info")

	return utilities.Ptr[cloudControllerResource.AllClusterData](v[0]), nil
}
