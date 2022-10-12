/*
Kubesimplify (c)
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
	ERRORCODE = "0x001"
	RegionLON = "LON1"
	RegionFRA = "FRA1"
	RegionNYC = "NYC1"
)

func getUserName() string {
	usrCmd := exec.Command("whoami")

	output, err := usrCmd.Output()
	if err != nil {
		return ""
	}
	userName := strings.Trim(string(output), "\n")
	return userName
}

func fetchAPIKey() string {

	file, err := os.ReadFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/cred/civo", getUserName()))
	if err != nil {
		return ""
	}
	return strings.Split(string(file), " ")[1]
}

func isValidRegion(reg string) bool {
	return strings.Compare(reg, RegionFRA) == 0 ||
		strings.Compare(reg, RegionNYC) == 0 ||
		strings.Compare(reg, RegionLON) == 0
}

func kubeconfigWriter(kubeconfig, clusterN, region, clusterID string) error {
	// create the neccessary folders and files
	err := os.Mkdir(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s", getUserName(), clusterN+"-"+region), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", getUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}
	_, err = os.Create(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", getUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}

	// write the contents to the req. files
	file, err := os.OpenFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", getUserName(), clusterN+"-"+region), os.O_WRONLY, 0755)
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
		fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", getUserName(), clusterN+"-"+region),
		[]byte(fmt.Sprintf("%s %s", clusterID, region)),
		0666)

	if err != nil {
		return err
	}

	//TODO: make this more reliable
	err = os.Setenv("KUBECONFIG", fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/config", getUserName(), clusterN+"-"+region))
	if err != nil {
		return err
	}
	return nil
}

func ClusterInfoInjecter(cName, reg, size string, nodes int) payload.CivoProvider {
	spec := payload.CivoProvider{
		ClusterName: cName,
		Region:      reg,
		APIKey:      fetchAPIKey(),
		HACluster:   false,
		Spec: payload.Machine{
			Nodes: nodes,
			Disk:  size,
		},
	}
	return spec
}

func CreateCluster(payload payload.CivoProvider) error {
	if len(payload.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !isValidRegion(payload.Region) {
		return fmt.Errorf("region code is Invalid")
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
	}

	resp, err := client.NewKubernetesClusters(configK8s)
	if err != nil {
		if errors.Is(err, civogo.DatabaseKubernetesClusterDuplicateError) {
			return fmt.Errorf("DUPLICATE NAME FOUND")
		}
		if errors.Is(err, civogo.AuthenticationFailedError) {
			return fmt.Errorf("AUTH FAILED")
		}
		if errors.Is(err, civogo.UnknownError) {
			return fmt.Errorf("UNKNOWN ERR")
		}
	}
	fmt.Println(resp.Status)
	//return resp.ID
	for true {
		clusterDS, _ := client.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {
			//print the new KUBECONFIG
			fmt.Println(clusterDS.KubeConfig)
			err := kubeconfigWriter(clusterDS.KubeConfig, payload.ClusterName, payload.Region, resp.ID)
			if err != nil {
				return err
			}
			break
		}
		fmt.Printf("Waiting.. Status: %v\n", clusterDS.Status)
		time.Sleep(15 * time.Second)
	}

	return nil
}

func kubeconfigDeleter(clustername, region string) error {
	err := os.RemoveAll(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s", getUserName(), clustername+"-"+region))
	if err != nil {
		return err
	}
	return nil
}

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
	// remove the KUBECONFIG and related configs
	return nil
}

func DeleteCluster(region, name string) error {
	data, err := os.ReadFile(fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/%s/info", getUserName(), name+"-"+region))
	if err != nil {
		return err
	}
	arr := strings.Split(string(data), " ")
	err = deleteClusterWithID(arr[1], arr[0])
	if err != nil {
		return err
	}
	return kubeconfigDeleter(name, region)
}
