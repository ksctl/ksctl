package civo

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/utils"

	"github.com/kubesimplify/ksctl/api/resources"
)

func watchManagedCluster(obj *CivoProvider, state resources.StateManagementInfrastructure, id string) error {

	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := civoClient.GetKubernetesCluster(id)
		if clusterDS.Ready {
			fmt.Println("[civo] Booted Instance", obj.Metadata.ResName)
			log.Default().Println(clusterDS.KubeConfig)
			civoCloudState.IsCompleted = true
			path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
			if err := saveStateHelper(state, path); err != nil {
				return err
			}
			path = generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
			err := saveKubeconfigHelper(state, path, clusterDS.KubeConfig)
			if err != nil {
				return err
			}
			break
		}
		fmt.Println("[civo] creating cluster..", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	return nil
}

// NewManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] creating managed %s cluster...", obj.Metadata.ResName)

	if len(civoCloudState.ManagedClusterID) != 0 {
		fmt.Println("[skip] managed cluster creation found", civoCloudState.ManagedClusterID)
		//check the state of creation via some watcher()

		if err := watchManagedCluster(obj, state, civoCloudState.ManagedClusterID); err != nil {
			return err
		}

		return nil // its a part of play back
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
		Applications:    "",
		CNIPlugin:       "cilium",
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
	if err := saveStateHelper(state, path); err != nil {
		return err
	}

	// save the stateID here so that we dont rcreate a new one
	// for {
	// 	// clusterDS fetches the current state of kubernetes cluster given its id
	// 	clusterDS, _ := civoClient.GetKubernetesCluster(resp.ID)
	// 	if clusterDS.Ready {
	// 		fmt.Println("[civo] Booted Instance", obj.Metadata.ResName)
	// 		log.Default().Println(clusterDS.KubeConfig)
	// 		civoCloudState.IsCompleted = true
	// 		path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
	// 		if err := saveStateHelper(state, path); err != nil {
	// 			return err
	// 		}
	// 		path = generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
	// 		err := saveKubeconfigHelper(state, path, clusterDS.KubeConfig)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		break
	// 	}
	// 	fmt.Println("[civo] creating cluster..", clusterDS.Status)
	// 	time.Sleep(10 * time.Second)
	// }
	if err := watchManagedCluster(obj, state, resp.ID); err != nil {
		return err
	}
	fmt.Println("[civo] Created your managed civo cluster!!🥳 🎉 ")
	return nil
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] Del Managed %s cluster....", obj.Metadata.ResName)
	if len(civoCloudState.ManagedClusterID) == 0 {
		fmt.Println("[skip] network deletion found")
		return nil
	}
	_, err := civoClient.DeleteKubernetesCluster(civoCloudState.ManagedClusterID)
	if err != nil {
		return err
	}
	civoCloudState.ManagedClusterID = ""
	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
	if err := saveStateHelper(state, path); err != nil {
		return err
	}

	return nil
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (obj *CivoProvider) GetManagedKubernetes(state resources.StateManagementInfrastructure) {
	fmt.Printf("[civo] Got Managed %s cluster....", obj.Metadata.ResName)
}
