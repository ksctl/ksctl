package civo

import (
	"fmt"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/civo/civogo"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func watchManagedCluster(obj *CivoProvider, storage types.StorageFactory, id string, name string) error {

	expoBackoff := helpers.NewBackOff(
		10*time.Second,
		2,
		2*int(consts.CounterMaxWatchRetryCount),
	)

	var clusterDS *civogo.KubernetesCluster
	_err := expoBackoff.Run(
		civoCtx,
		log,
		func() (err error) {
			clusterDS, err = obj.client.GetKubernetesCluster(id)
			return err
		},
		func() bool {
			return clusterDS.Ready
		},
		nil,
		func() error {
			log.Print(civoCtx, "cluster ready", "name", name)
			mainStateDocument.CloudInfra.Civo.B.IsCompleted = true
			mainStateDocument.ClusterKubeConfig = clusterDS.KubeConfig
			mainStateDocument.ClusterKubeConfigContext = name
			return storage.Write(mainStateDocument)
		},
		fmt.Sprintf("Waiting for managed cluster %s to be ready", id),
	)
	if _err != nil {
		return _err
	}

	return nil
}

func (obj *CivoProvider) NewManagedCluster(storage types.StorageFactory, noOfNodes int) error {

	name := <-obj.chResName
	vmtype := <-obj.chVMType

	log.Debug(civoCtx, "Printing", "name", name, "vmtype", vmtype)

	if len(mainStateDocument.CloudInfra.Civo.ManagedClusterID) != 0 {
		log.Print(civoCtx, "skipped managed cluster creation found", "id", mainStateDocument.CloudInfra.Civo.ManagedClusterID)

		if err := watchManagedCluster(obj, storage, mainStateDocument.CloudInfra.Civo.ManagedClusterID, name); err != nil {
			return err
		}

		return nil
	}

	configK8s := &civogo.KubernetesClusterConfig{
		KubernetesVersion: obj.metadata.k8sVersion,
		Name:              name,
		Region:            obj.region,
		NumTargetNodes:    noOfNodes,
		TargetNodesSize:   vmtype,
		NetworkID:         mainStateDocument.CloudInfra.Civo.NetworkID,
		Applications:      obj.metadata.apps, // make the use of application and cni via some method
		CNIPlugin:         obj.metadata.cni,  // make it use install application in the civo
	}
	log.Debug(civoCtx, "Printing", "configManagedK8s", configK8s)

	resp, err := obj.client.NewKubernetesClusters(configK8s)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Civo.NoManagedNodes = noOfNodes
	mainStateDocument.BootstrapProvider = "managed"
	mainStateDocument.CloudInfra.Civo.ManagedNodeSize = vmtype
	mainStateDocument.CloudInfra.Civo.B.KubernetesVer = obj.metadata.k8sVersion
	mainStateDocument.CloudInfra.Civo.ManagedClusterID = resp.ID

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	if err := watchManagedCluster(obj, storage, resp.ID, name); err != nil {
		return err
	}
	log.Success(civoCtx, "Created Managed cluster", "clusterID", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	return nil
}

func (obj *CivoProvider) DelManagedCluster(storage types.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Civo.ManagedClusterID) == 0 {
		log.Print(civoCtx, "skipped network deletion found", "id", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
		return nil
	}
	_, err := obj.client.DeleteKubernetesCluster(mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	if err != nil {
		return err
	}
	log.Success(civoCtx, "Deleted Managed cluster", "clusterID", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	mainStateDocument.CloudInfra.Civo.ManagedClusterID = ""
	mainStateDocument.CloudInfra.Civo.ManagedNodeSize = ""

	return storage.Write(mainStateDocument)
}
