// Copyright 2024 ksctl
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

//
//import (
//	"context"
//
//	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
//	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
//	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"
//	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"
//	bootstrapController "github.com/ksctl/ksctl/pkg/controllers/bootstrap"
//	cloudController "github.com/ksctl/ksctl/pkg/controllers/cloud"
//	"github.com/ksctl/ksctl/pkg/helpers"
//	"github.com/ksctl/ksctl/pkg/helpers/consts"
//	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
//	"github.com/ksctl/ksctl/pkg/helpers/utilities"
//	"github.com/ksctl/ksctl/pkg/types"
//	cloudControllerResource "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
//	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
//)
//
//type ManagerClusterKsctl struct {
//	managerInfo
//}
//
//func NewManagerClusterKsctl(ctx context.Context, log types.LoggerFactory, client *types.KsctlClient) (*ManagerClusterKsctl, error) {
//	defer panicCatcher(log)
//
//	stateDocument = new(storageTypes.StorageDocument)
//	controllerCtx = context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl-manager")
//
//	cloudController.InitLogger(controllerCtx, log)
//	bootstrapController.InitLogger(controllerCtx, log)
//
//	manager := new(ManagerClusterKsctl)
//	manager.log = log
//	manager.client = client
//
//	if err := manager.initStorage(controllerCtx); err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	if err := manager.startPoller(controllerCtx); err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	return manager, nil
//}
//
//func (manager *ManagerClusterKsctl) Credentials() error {
//	log := manager.log
//	client := manager.client
//
//	defer panicCatcher(log)
//
//	var err error
//	switch client.Metadata.Provider {
//	case consts.CloudCivo:
//		client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)
//
//	case consts.CloudAzure:
//		client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)
//
//	case consts.CloudAws:
//		client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)
//
//	default:
//		err = ksctlErrors.ErrInvalidCloudProvider.Wrap(
//			manager.log.NewError(
//				controllerCtx, "Problem in validation", "cloud", client.Metadata.Provider,
//			),
//		)
//	}
//
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	err = client.Cloud.Credential(client.Storage)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//	log.Success(controllerCtx, "Successfully Credential Added")
//
//	return nil
//}
//
//func (manager *ManagerClusterKsctl) SwitchCluster() (*string, error) {
//	client := manager.client
//	log := manager.log
//	defer panicCatcher(log)
//
//	if err := manager.setupConfigurations(); err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	if client.Metadata.Provider == consts.CloudLocal {
//		client.Metadata.Region = "LOCAL"
//	}
//
//	clusterType := consts.ClusterTypeMang
//	if client.Metadata.IsHA {
//		clusterType = consts.ClusterTypeHa
//	}
//
//	if err := client.Storage.Setup(
//		client.Metadata.Provider,
//		client.Metadata.Region,
//		client.Metadata.ClusterName,
//		clusterType); err != nil {
//
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	defer func() {
//		if err := client.Storage.Kill(); err != nil {
//			log.Error("StorageClass Kill failed", "reason", err)
//		}
//	}()
//
//	var err error
//	switch client.Metadata.Provider {
//	case consts.CloudCivo:
//		client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, civoPkg.ProvideClient)
//
//	case consts.CloudAzure:
//		client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, azurePkg.ProvideClient)
//
//	case consts.CloudAws:
//		client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, awsPkg.ProvideClient)
//		if err != nil {
//			break
//		}
//
//		err = cloudController.InitCloud(client, stateDocument, consts.OperationGet)
//
//	case consts.CloudLocal:
//		client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, stateDocument, localPkg.ProvideClient)
//
//	}
//
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	if err := client.Cloud.IsPresent(client.Storage); err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	kubeconfig, err := client.Cloud.GetKubeconfig(client.Storage)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	} else {
//		if kubeconfig == nil {
//			err = ksctlErrors.ErrKubeconfigOperations.Wrap(
//				manager.log.NewError(
//					controllerCtx, "Problem in kubeconfig get"),
//			)
//
//			log.Error("Kubeconfig we got is nil")
//			return nil, err
//		}
//	}
//
//	path, err := helpers.WriteKubeConfig(controllerCtx, *kubeconfig)
//	log.Debug(controllerCtx, "data", "kubeconfigPath", path)
//
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	printKubeConfig(manager.log, path)
//
//	return kubeconfig, nil
//}
//
//func (manager *ManagerClusterKsctl) clusterDataHelper(
//	operation consts.LogClusterDetail) ([]cloudControllerResource.AllClusterData, error) {
//
//	client := manager.client
//	log := manager.log
//	defer panicCatcher(log)
//
//	if err := manager.validationFields(client.Metadata); err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	if client.Metadata.Provider == consts.CloudLocal {
//		client.Metadata.Region = "LOCAL"
//	}
//
//	if operation == consts.LoggingInfoCluster {
//		var err []error
//		if len(client.Metadata.ClusterName) == 0 {
//			err = append(err,
//				ksctlErrors.ErrInvalidResourceName.Wrap(
//					log.NewError(controllerCtx, "clustername is needed for cluster info details"),
//				),
//			)
//		}
//		if len(client.Metadata.Region) == 0 {
//			err = append(err,
//				ksctlErrors.ErrInvalidCloudRegion.Wrap(
//					log.NewError(controllerCtx, "region is needed for cluster info details"),
//				),
//			)
//		}
//		if len(err) != 0 {
//			_err := ksctlErrors.ErrInvalidUserInput.Wrap(
//				log.NewError(controllerCtx, "Failure", "reason", err),
//			)
//
//			log.Error("Failure", "reason", _err)
//
//			return nil, _err
//		}
//	}
//
//	defer func() {
//		if err := client.Storage.Kill(); err != nil {
//			log.Error("StorageClass Kill failed", "reason", err)
//		}
//	}()
//
//	log.Note(controllerCtx, "Filter", "cloudProvider", string(client.Metadata.Provider))
//
//	var (
//		cloudMapper = map[consts.KsctlCloud]types.CloudFactory{
//			consts.CloudCivo:  nil,
//			consts.CloudAzure: nil,
//			consts.CloudAws:   nil,
//			consts.CloudLocal: nil,
//		}
//	)
//
//	var err error
//	switch client.Metadata.Provider {
//	case consts.CloudCivo:
//		cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)
//
//	case consts.CloudAzure:
//		cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)
//
//	case consts.CloudAws:
//		cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)
//
//	case consts.CloudLocal:
//		cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)
//
//	default:
//		switch operation {
//		case consts.LoggingGetClusters:
//			if client.Metadata.Provider != consts.CloudAll {
//				err = ksctlErrors.ErrInvalidCloudProvider.Wrap(
//					manager.log.NewError(
//						controllerCtx, "", "cloud", client.Metadata.Provider,
//					),
//				)
//			} else {
//				cloudMapper[consts.CloudCivo], err = civoPkg.NewClient(controllerCtx, client.Metadata, log, nil, civoPkg.ProvideClient)
//				if err != nil {
//					log.Error("handled error", "catch", err)
//					return nil, err
//				}
//				cloudMapper[consts.CloudAzure], err = azurePkg.NewClient(controllerCtx, client.Metadata, log, nil, azurePkg.ProvideClient)
//				if err != nil {
//					log.Error("handled error", "catch", err)
//					return nil, err
//				}
//				cloudMapper[consts.CloudAws], err = awsPkg.NewClient(controllerCtx, client.Metadata, log, nil, awsPkg.ProvideClient)
//				if err != nil {
//					log.Error("handled error", "catch", err)
//					return nil, err
//				}
//				cloudMapper[consts.CloudLocal], err = localPkg.NewClient(controllerCtx, client.Metadata, log, nil, localPkg.ProvideClient)
//				if err != nil {
//					log.Error("handled error", "catch", err)
//					return nil, err
//				}
//			}
//
//		case consts.LoggingInfoCluster:
//			err = ksctlErrors.ErrInvalidCloudProvider.Wrap(
//				manager.log.NewError(
//					controllerCtx, "", "cloud", client.Metadata.Provider,
//				),
//			)
//		}
//
//	}
//
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return nil, err
//	}
//
//	var printerTable []cloudControllerResource.AllClusterData
//	for _, v := range cloudMapper {
//		if v == nil {
//			continue
//		}
//		data, err := v.GetRAWClusterInfos(client.Storage)
//		if err != nil {
//			log.Error("handled error", "catch", err)
//			return nil, err
//		}
//
//		if operation == consts.LoggingInfoCluster {
//			// as the info will not have all as cloud provider this loop will only run once
//			for _, _data := range data {
//				if _data.Name != manager.client.Metadata.ClusterName ||
//					_data.Region != manager.client.Metadata.Region {
//					continue
//				} else {
//					printerTable = append(printerTable, _data)
//					return printerTable, nil
//				}
//			}
//			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
//				log.NewError(controllerCtx, "No state is present",
//					"name", manager.client.Metadata.ClusterName),
//			)
//		}
//
//		printerTable = append(printerTable, data...)
//	}
//	return printerTable, err
//}
//
//func (manager *ManagerClusterKsctl) GetCluster() error {
//	v, err := manager.clusterDataHelper(
//		consts.LoggingGetClusters,
//	)
//	if err != nil {
//		return err
//	}
//
//	manager.log.Table(controllerCtx, consts.LoggingGetClusters, v)
//
//	manager.log.Success(controllerCtx, "successfully get clusters")
//
//	return nil
//}
//
//func (manager *ManagerClusterKsctl) InfoCluster() (*cloudControllerResource.AllClusterData, error) {
//	v, err := manager.clusterDataHelper(
//		consts.LoggingInfoCluster,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	manager.log.Table(controllerCtx, consts.LoggingInfoCluster, v)
//
//	manager.log.Success(controllerCtx, "successfully cluster info")
//
//	return utilities.Ptr[cloudControllerResource.AllClusterData](v[0]), nil
//}
