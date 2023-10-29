package civo

import (
	"errors"
	"fmt"
	"time"

	"github.com/civo/civogo"

	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func watchManagedCluster(obj *CivoProvider, storage resources.StorageFactory, id string, name string) error {

	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		//NOTE: this is prone to network failure
		var clusterDS *civogo.KubernetesCluster
		currRetryCounter := KsctlCounterConsts(0)
		for currRetryCounter < CounterMaxWatchRetryCount {
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
		if currRetryCounter == CounterMaxWatchRetryCount {
			return fmt.Errorf("[civo] failed to get the state of managed cluster")
		}

		if clusterDS.Ready {
			fmt.Println("[civo] Booted Instance", name)
			civoCloudState.IsCompleted = true
			path := generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)
			if err := saveStateHelper(storage, path); err != nil {
				return err
			}
			path = generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
			err := saveKubeconfigHelper(storage, path, clusterDS.KubeConfig)
			if err != nil {
				return err
			}
			printKubeconfig(storage, OperationStateCreate)
			break
		}
		log.Debug("creating cluster..", "name", name, "Status", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	name := obj.metadata.resName
	vmtype := obj.metadata.vmType
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	log.Debug("Printing", "name", name, "vmtype", vmtype)

	if len(civoCloudState.ManagedClusterID) != 0 {
		log.Print("skipped managed cluster creation found", civoCloudState.ManagedClusterID)

		if err := watchManagedCluster(obj, storage, civoCloudState.ManagedClusterID, name); err != nil {
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
		NetworkID:         civoCloudState.NetworkIDs.NetworkID,
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

	civoCloudState.NoManagedNodes = noOfNodes
	civoCloudState.KubernetesDistro = string(K8sK3s)
	civoCloudState.KubernetesVer = obj.metadata.k8sVersion
	civoCloudState.ManagedClusterID = resp.ID

	path := generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)
	if err := saveStateHelper(storage, path); err != nil {
		return log.NewError(err.Error())
	}

	if err := watchManagedCluster(obj, storage, resp.ID, name); err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Created Managed cluster", "clusterID", civoCloudState.ManagedClusterID)
	return nil
}

// DelManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(civoCloudState.ManagedClusterID) == 0 {
		log.Print("skipped network deletion found", "id", civoCloudState.ManagedClusterID)
		return nil
	}
	_, err := obj.client.DeleteKubernetesCluster(civoCloudState.ManagedClusterID)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("Deleted Managed cluster", "clusterID", civoCloudState.ManagedClusterID)
	civoCloudState.ManagedClusterID = ""
	path := generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)
	if err := saveStateHelper(storage, path); err != nil {
		return log.NewError(err.Error())
	}
	printKubeconfig(storage, OperationStateDelete)

	return nil
}

func (obj *CivoProvider) GetKubeconfigPath() string {
	return generatePath(UtilClusterPath, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
}
