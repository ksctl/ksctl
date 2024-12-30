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

//import (
//	"strings"
//
//	"github.com/ksctl/ksctl/pkg/consts"
//	cloudControllerResource "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
//)
//
//func (kc *Controller) AddWorkerPlaneNodes() error {
//
//	if !client.Metadata.IsHA {
//		err := log.NewError(controllerCtx, "this feature is only for ha clusters")
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	if err := client.Storage.Setup(
//		client.Metadata.Provider,
//		client.Metadata.Region,
//		client.Metadata.ClusterName,
//		consts.ClusterTypeHa); err != nil {
//
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	defer func() {
//		if err := client.Storage.Kill(); err != nil {
//			log.Error("StorageClass Kill failed", "reason", err)
//		}
//	}()
//
//	if err := cloudController.InitCloud(client, stateDocument, consts.OperationGet); err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	err := bootstrapController.Setup(client, stateDocument)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	currWP, cloudResErr := cloudController.AddWorkerNodes(client)
//	if cloudResErr != nil {
//		log.Error("handled error", "catch", cloudResErr)
//		return cloudResErr
//	}
//
//	var payload cloudControllerResource.CloudResourceState
//	payload, err = client.Cloud.GetStateForHACluster(client.Storage)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	err = bootstrapController.JoinMoreWorkerPlanes(client, currWP, client.Metadata.NoWP)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	log.Success(controllerCtx, "successfully added workernodes")
//	return nil
//}
//
//func (kc *Controller) DelWorkerPlaneNodes() error {
//
//	client := kc.client
//	log := kc.log
//	defer panicCatcher(log)
//
//	if err := kc.setupConfigurations(); err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	if !client.Metadata.IsHA {
//		err := log.NewError(controllerCtx, "this feature is only for ha clusters")
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	if err := client.Storage.Setup(
//		client.Metadata.Provider,
//		client.Metadata.Region,
//		client.Metadata.ClusterName,
//		consts.ClusterTypeHa); err != nil {
//
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	defer func() {
//		if err := client.Storage.Kill(); err != nil {
//			log.Error("handled error", "catch", err)
//			log.Error("StorageClass Kill failed", "reason", err)
//		}
//	}()
//
//	fakeClient := false
//	if _, ok := helpers.IsContextPresent(controllerCtx, consts.KsctlTestFlagKey); ok {
//		fakeClient = true
//	}
//
//	if err := cloudController.InitCloud(client, stateDocument, consts.OperationGet); err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	err := bootstrapController.Setup(client, stateDocument)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	hostnames, err := cloudController.DelWorkerNodes(client)
//	if err != nil {
//		log.Error("handled error", "catch", err)
//		return err
//	}
//
//	log.Debug(controllerCtx, "K8s nodes to be deleted", "hostnames", strings.Join(hostnames, ";"))
//
//	if !fakeClient {
//		var payload cloudControllerResource.CloudResourceState
//		payload, err = client.Cloud.GetStateForHACluster(client.Storage)
//		if err != nil {
//			log.Error("handled error", "catch", err)
//			return err
//		}
//
//		err = client.PreBootstrap.Setup(payload, client.Storage, consts.OperationGet)
//		if err != nil {
//			log.Error("handled error", "catch", err)
//			return err
//		}
//
//		if err := bootstrapController.DelWorkerPlanes(
//			client,
//			stateDocument.ClusterKubeConfig,
//			hostnames); err != nil {
//
//			log.Error("handled error", "catch", err)
//			return err
//		}
//	}
//	log.Success(controllerCtx, "Successfully deleted workerNodes")
//
//	return nil
//}
