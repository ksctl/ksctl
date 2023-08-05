package civo

import (
	"errors"
	"fmt"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/kubesimplify/ksctl/api/resources"
)

func watchManagedCluster(obj *CivoProvider, storage resources.StateManagementInfrastructure, id string) error {

	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := civoClient.GetKubernetesCluster(id)
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

// NewManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewManagedCluster(storage resources.StateManagementInfrastructure) error {

	if len(civoCloudState.ManagedClusterID) != 0 {
		fmt.Println("[skip] managed cluster creation found", civoCloudState.ManagedClusterID)

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
		Name:            obj.Metadata.ResName,
		Region:          obj.Region,
		NumTargetNodes:  obj.NoOfManagedNodes,
		TargetNodesSize: obj.Metadata.VmType,
		NetworkID:       network.ID,
		Applications:    "",       // make the use of application and cni via some method
		CNIPlugin:       "cilium", // make it use install application in the civo
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
	}
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

// DelManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelManagedCluster(storage resources.StateManagementInfrastructure) error {
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

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (obj *CivoProvider) GetManagedKubernetes(storage resources.StateManagementInfrastructure) {
	// TODO: used for getting information on all the clusters created with all the types
	// ha and managed in some form of predefined json format
	fmt.Printf("[civo] Got Managed %s cluster....", obj.Metadata.ResName)
}
