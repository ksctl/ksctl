/*
Kubesimplify (c)
Credit to @civo
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package civo

import (
	"errors"
	"fmt"
	"github.com/civo/civogo"
	"github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	RegionLON = "LON1"
	RegionFRA = "FRA1"
	RegionNYC = "NYC1"
)

// getUserName returns current active username
func getUserName() string {
	usrCmd := exec.Command("whoami")

	output, err := usrCmd.Output()
	if err != nil {
		return ""
	}
	userName := strings.Trim(string(output), "\n")
	return userName
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/cred/civo", payload.GetUserName()))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

// isValidRegion Checks whether the Region passed by user is valid according to CIVO
func isValidRegion(reg string) bool {
	return strings.Compare(reg, RegionFRA) == 0 ||
		strings.Compare(reg, RegionNYC) == 0 ||
		strings.Compare(reg, RegionLON) == 0
}

//TODO: Refactoring of fmt.Sprintf statements

// kubeconfigWriter Writes kubeconfig supplied to config directory of respective cluster created
func kubeconfigWriter(kubeconfig, clusterN, region, clusterID string) error {
	// create the neccessary folders and files
	err := os.Mkdir(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s", payload.GetUserName(), clusterN+"-"+region), 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", payload.GetUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}
	_, err = os.Create(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", payload.GetUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}

	// write the contents to the req. files
	file, err := os.OpenFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", payload.GetUserName(), clusterN+"-"+region), os.O_WRONLY, 0750)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(kubeconfig))
	if err != nil {
		return err
	}

	// write the info of cluster
	// @FORMAT: | ClusterID Region |
	err = os.WriteFile(
		fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", payload.GetUserName(), clusterN+"-"+region),
		[]byte(fmt.Sprintf("%s %s", clusterID, region)),
		0640)

	if err != nil {
		return err
	}

	//FIXME: make this more reliable
	err = os.Setenv("KUBECONFIG", fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", payload.GetUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}
	return nil
}

// ClusterInfoInjecter Serializes the information which is return as payload.CivoProvider for sending it to API
// {clustername, regionCode, Size of Nodes, No of nodes, Applications(optional), cniPlugion(optional)}
func ClusterInfoInjecter(clusterName, reg, size string, noOfNodes int, application, cniPlugin string) payload.CivoProvider {

	if len(application) == 0 {
		application = "Traefik-v2-nodeport,metrics-server" // default: applications
	} else {
		application += ",Traefik-v2-nodeport,metrics-server"
	}
	if len(cniPlugin) == 0 {
		cniPlugin = "flannel" // default: flannel
	}

	spec := payload.CivoProvider{
		ClusterName: clusterName,
		Region:      reg,
		APIKey:      fetchAPIKey(),
		HACluster:   false,
		Spec: payload.Machine{
			Nodes: noOfNodes,
			Disk:  size,
			Mem:   "0M",
			Cpu:   "1m",
		},
		Application: application,
		CNIPlugin:   cniPlugin,
	}
	return spec
}

// isPresent Checks whether the cluster to create is already had been created
func isPresent(clusterName, Region string) bool {
	_, err := os.ReadFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", payload.GetUserName(), clusterName+"-"+Region))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// isValidSize checks whether the size passed by user is valid according to CIVO
func isValidSize(size string) bool {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	for _, valid := range validSizes {
		if strings.Compare(valid, size) == 0 {
			return true
		}
	}
	return false
}

// CreateCluster creates cluster as provided configuration and returns whether it fails or not
func CreateCluster(payload payload.CivoProvider) error {
	if len(payload.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !isValidRegion(payload.Region) {
		return fmt.Errorf("region code is Invalid")
	}

	if isPresent(payload.ClusterName, payload.Region) {
		return fmt.Errorf("DUPLICATE Cluster")
	}

	if !isValidSize(payload.Spec.Disk) {
		return fmt.Errorf("INVALID size of node")
	}

	client, err := civogo.NewClient(payload.APIKey, payload.Region)
	if err != nil {
		return err
	}

	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		return err
	}

	configK8s := &civogo.KubernetesClusterConfig{
		Name:            payload.ClusterName,
		Region:          payload.Region,
		NumTargetNodes:  payload.Spec.Nodes,
		TargetNodesSize: payload.Spec.Disk,
		NetworkID:       defaultNetwork.ID,
		Applications:    payload.Application,
		CNIPlugin:       payload.CNIPlugin,
	}

	resp, err := client.NewKubernetesClusters(configK8s)
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
	fmt.Println(resp.Status)
	for true {
		clusterDS, _ := client.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {
			//print the new KUBECONFIG
			err := kubeconfigWriter(clusterDS.KubeConfig, payload.ClusterName, payload.Region, resp.ID)
			if err != nil {
				return err
			}

			fmt.Println("Do you want to print KUBECONFIG here?[y/N]: ")
			choice := byte(' ')
			_, err = fmt.Scanf("%c", &choice)

			if err != nil {
				return err
			}

			if choice == 'y' || choice == 'Y' {
				fmt.Println("########################")
				fmt.Println(clusterDS.KubeConfig)
				fmt.Println("########################")
			}

			break
		}
		fmt.Printf("Waiting.. Status: %v\n", clusterDS.Status)
		time.Sleep(15 * time.Second)
	}

	return nil
}

//kubeconfigDeleter deletes all configs related to the provided cluster
func kubeconfigDeleter(clustername, region string) error {
	err := os.RemoveAll(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s", payload.GetUserName(), clustername+"-"+region))
	if err != nil {
		return err
	}
	return nil
}

// deleteClusterWithID delete cluster from CIVO by provided regionCode and clusterID
func deleteClusterWithID(regionCode string, clusterID string) error {
	client, err := civogo.NewClient(fetchAPIKey(), regionCode)
	if err != nil {
		return err
	}

	cluster, err := client.DeleteKubernetesCluster(clusterID)
	if err != nil {
		return err
	}
	fmt.Println(string(cluster.Result))
	return nil
}

// DeleteCluster deletes cluster from the given name and region
func DeleteCluster(region, name string) error {
	data, err := os.ReadFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", payload.GetUserName(), name+"-"+region))
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}
	arr := strings.Split(string(data), " ")
	err = deleteClusterWithID(arr[1], arr[0])
	if err != nil {
		return err
	}
	return kubeconfigDeleter(name, region)
}
