package civo

import (
	"errors"
	"fmt"
	"time"

	"github.com/civo/civogo"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func watchManagedCluster(obj *CivoProvider, storage resources.StorageFactory, id string, name string) error {

	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		//NOTE: this is prone to network failure
		var clusterDS *civogo.KubernetesCluster
		currRetryCounter := consts.KsctlCounterConsts(0)
		for currRetryCounter < consts.CounterMaxWatchRetryCount {
			var err error
			clusterDS, err = obj.client.GetKubernetesCluster(id)
			if err != nil {
				currRetryCounter++
				log.Warn("RETRYING", err)
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == consts.CounterMaxWatchRetryCount {
			return fmt.Errorf("[civo] failed to get the state of managed cluster")
		}

		if clusterDS.Ready {
			fmt.Println("[civo] Booted Instance", name)
			mainStateDocument.CloudInfra.Civo.B.IsCompleted = true
			mainStateDocument.ClusterKubeConfig = clusterDS.KubeConfig
			err := storage.Write(mainStateDocument)
			if err != nil {
				return err
			}
			break
		}
		log.Debug("creating cluster..", "name", name, "Status", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	name := <-obj.chResName
	vmtype := <-obj.chVMType

	log.Debug("Printing", "name", name, "vmtype", vmtype)

	if len(mainStateDocument.CloudInfra.Civo.ManagedClusterID) != 0 {
		log.Print("skipped managed cluster creation found", mainStateDocument.CloudInfra.Civo.ManagedClusterID)

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
	log.Debug("Printing", "configManagedK8s", configK8s)

	resp, err := obj.client.NewKubernetesClusters(configK8s)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			return log.NewError("DUPLICATE Cluster FOUND")
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			return log.NewError("AUTH FAILED")
		}
		if errors.Is(err, civogo.UnknownError) {
			return log.NewError("UNKNOWN ERR")
		}
		return log.NewError(err.Error())
	}

	mainStateDocument.CloudInfra.Civo.NoManagedNodes = noOfNodes
	mainStateDocument.CloudInfra.Civo.B.KubernetesDistro = string(consts.K8sK3s)
	mainStateDocument.CloudInfra.Civo.B.KubernetesVer = obj.metadata.k8sVersion
	mainStateDocument.CloudInfra.Civo.ManagedClusterID = resp.ID

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	if err := watchManagedCluster(obj, storage, resp.ID, name); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Created Managed cluster", "clusterID", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	return nil
}

// DelManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Civo.ManagedClusterID) == 0 {
		log.Print("skipped network deletion found", "id", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
		return nil
	}
	_, err := obj.client.DeleteKubernetesCluster(mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Deleted Managed cluster", "clusterID", mainStateDocument.CloudInfra.Civo.ManagedClusterID)
	mainStateDocument.CloudInfra.Civo.ManagedClusterID = ""

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}

	return nil
}
