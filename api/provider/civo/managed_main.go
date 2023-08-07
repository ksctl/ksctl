package civo

import (
	"errors"
	"fmt"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/kubesimplify/ksctl/api/resources"
)

func watchManagedCluster(obj *CivoProvider, storage resources.StorageFactory, id string) error {

	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		//NOTE: this is prone to network failure
		var clusterDS *civogo.KubernetesCluster
		currRetryCounter := 0
		for currRetryCounter < utils.MAX_RETRY_COUNT {
			var err error
			clusterDS, err = civoClient.GetKubernetesCluster(id)
			if err != nil {
				currRetryCounter++
				storage.Logger().Err(fmt.Sprintln("RETRYING", err))
			} else {
				return err
			}
			time.Sleep(5 * time.Second)
		}

		if clusterDS.Ready {
			fmt.Println("[civo] Booted Instance", obj.Metadata.ResName)
			civoCloudState.IsCompleted = true
			path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
			if err := saveStateHelper(storage, path); err != nil {
				return err
			}
			path = generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
			err := saveKubeconfigHelper(storage, path, clusterDS.KubeConfig)
			if err != nil {
				return err
			}
			printKubeconfig(storage, "create")
			break
		}
		storage.Logger().Print("[civo] creating cluster..", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	if len(civoCloudState.ManagedClusterID) != 0 {
		storage.Logger().Success("[skip] managed cluster creation found", civoCloudState.ManagedClusterID)

		if err := watchManagedCluster(obj, storage, civoCloudState.ManagedClusterID); err != nil {
			return err
		}

		return nil
	}

	network, err := civoClient.GetNetwork(civoCloudState.NetworkIDs.NetworkID)
	if err != nil {
		return err
	}

	configK8s := &civogo.KubernetesClusterConfig{
		KubernetesVersion: obj.Metadata.Version,
		Name:              obj.Metadata.ResName,
		Region:            obj.Region,
		NumTargetNodes:    noOfNodes,
		TargetNodesSize:   obj.Metadata.VmType,
		NetworkID:         network.ID,
		Applications:      obj.Metadata.Apps, // make the use of application and cni via some method
		CNIPlugin:         obj.Metadata.Cni,  // make it use install application in the civo
	}
	resp, err := civoClient.NewKubernetesClusters(configK8s)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			return fmt.Errorf("DUPLICATE Cluster FOUND")
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			return fmt.Errorf("AUTH FAILED")
		}
		if errors.Is(err, civogo.UnknownError) {
			return fmt.Errorf("UNKNOWN ERR")
		}
		return err
	}

	civoCloudState.NoManagedNodes = noOfNodes
	civoCloudState.KubernetesDistro = "k3s"
	civoCloudState.KubernetesVer = obj.Metadata.Version
	civoCloudState.ManagedClusterID = resp.ID

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
	if err := saveStateHelper(storage, path); err != nil {
		return err
	}

	if err := watchManagedCluster(obj, storage, resp.ID); err != nil {
		return err
	}
	return nil
}

// DelManagedCluster implements resources.CloudFactory.
func (obj *CivoProvider) DelManagedCluster(storage resources.StorageFactory) error {
	if len(civoCloudState.ManagedClusterID) == 0 {
		storage.Logger().Success("[skip] network deletion found")
		return nil
	}
	_, err := civoClient.DeleteKubernetesCluster(civoCloudState.ManagedClusterID)
	if err != nil {
		return err
	}
	storage.Logger().Success("[civo] Deleted Managed cluster", civoCloudState.ManagedClusterID)
	civoCloudState.ManagedClusterID = ""
	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
	if err := saveStateHelper(storage, path); err != nil {
		return err
	}
	printKubeconfig(storage, "delete")

	return nil
}

// GetManagedKubernetes implements resources.CloudFactory.
func (obj *CivoProvider) GetManagedKubernetes(storage resources.StorageFactory) {
	// TODO: used for getting information on all the clusters created with all the types
	// ha and managed in some form of predefined json format
	fmt.Printf("[civo] Got Managed %s cluster....", obj.Metadata.ResName)
}
