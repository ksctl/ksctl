package civo

import (
	"errors"
	"fmt"
	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/utils"
	"log"
	"time"

	"github.com/kubesimplify/ksctl/api/resources"
)

// NewManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] creating managed %s cluster...", obj.Metadata.ResName)
	// TODO: have a validation
	if err := validationOfArguments(obj.Metadata.ResName, obj.Region); err != nil {
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
	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := civoClient.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {
			fmt.Println("[civo] Booted Instance", obj.Metadata.ResName)
			log.Default().Println(clusterDS.KubeConfig)
			path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)
			if err := saveStateHelper(state, path); err != nil {
				return err
				//failed to save the ID
				// cleanup
			}
			path = generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
			err := saveKubeconfigHelper(state, path, clusterDS.KubeConfig)
			if err != nil {
				return err
			}
			break
		}
		fmt.Println("🚧 Instance", clusterDS.Status)
		time.Sleep(10 * time.Second)
	}
	fmt.Println("Created your managed civo cluster!!🥳 🎉 ")
	return nil
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] Del Managed %s cluster....", obj.Metadata.ResName)
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
