/*
Kubesimplify (c)
Credit to @civo
@maintainer: Dipankar Das <dipankardas0115@gmail.com> , Anurag Kumar <contact.anurag7@gmail.com>
*/

package civo

import (
	"errors"
	"fmt"
	"github.com/kubesimplify/Kubesimpctl/src/api/payload"
	"os"
	"strings"
	"time"

	"github.com/civo/civogo"
)

const (
	RegionLON = "LON1"
	RegionFRA = "FRA1"
	RegionNYC = "NYC1"
)

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

var (
	kubeconfig = fmt.Sprintf("/home/%s/.kube/kubesimpctl/config/civo/", payload.GetUserName())
)

// kubeconfigWriter Writes kubeconfig supplied to config directory of respective cluster created
func kubeconfigWriter(kubeconfig, clusterN, region, clusterID string) error {
	// create the neccessary folders and files
	workingDir := kubeconfig + clusterN + "-" + region
	err := os.Mkdir(workingDir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(workingDir + "/config")
	if err != nil {
		return err
	}
	_, err = os.Create(workingDir + "/info")
	if err != nil {
		return err
	}

	// write the contents to the req. files
	file, err := os.OpenFile(workingDir+"/config", os.O_WRONLY, 0750)
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
	if err = os.WriteFile(workingDir+"/info",
		[]byte(fmt.Sprintf("%s %s", clusterID, region)), 0640); err != nil {
		return err
	}

	//FIXME: make this more reliable
	err = os.Setenv("KUBECONFIG", workingDir+"/config")
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
	_, err := os.ReadFile(kubeconfig + clusterName + "-" + Region + "/info")
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

type printer struct {
	ClusterName string
	Region      string
}

func (p printer) Printer(a int) {
	switch a {
	case 0:
		fmt.Printf("\nTo use this cluster set this environment variable\n\n")
		fmt.Println(fmt.Sprintf("export KUBECONFIG='/home/%s/.kube/kubesimpctl/config/civo/%s/config'", payload.GetUserName(), p.ClusterName+"-"+p.Region))
	case 1:
		fmt.Printf("\nUse the following command to unset KUBECONFIG\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG"))
	}
	fmt.Println()
}

// CreateCluster creates cluster as provided configuration and returns whether it fails or not
func CreateCluster(cargo payload.CivoProvider) error {
	if len(cargo.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !isValidRegion(cargo.Region) {
		return fmt.Errorf("region code is Invalid")
	}

	if isPresent(cargo.ClusterName, cargo.Region) {
		return fmt.Errorf("DUPLICATE Cluster")
	}

	if !isValidSize(cargo.Spec.Disk) {
		return fmt.Errorf("INVALID size of node")
	}

	client, err := civogo.NewClient(cargo.APIKey, cargo.Region)
	if err != nil {
		return err
	}

	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		return err
	}

	configK8s := &civogo.KubernetesClusterConfig{
		Name:            cargo.ClusterName,
		Region:          cargo.Region,
		NumTargetNodes:  cargo.Spec.Nodes,
		TargetNodesSize: cargo.Spec.Disk,
		NetworkID:       defaultNetwork.ID,
		Applications:    cargo.Application,
		CNIPlugin:       cargo.CNIPlugin,
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
	for true {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := client.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {

			err := kubeconfigWriter(clusterDS.KubeConfig, cargo.ClusterName, cargo.Region, resp.ID)
			if err != nil {
				return err
			}

			fmt.Println("Do you want to print KUBECONFIG here?[y/N]: ")
			choice := byte(' ')
			_, err = fmt.Scanf("%c", &choice)

			if err != nil {
				return err
			}

			switch choice {
			case 'y', 'Y':
				fmt.Printf("\nHere is your KUBECONFIG\n\n")
				fmt.Println(clusterDS.KubeConfig)
				fmt.Println()
			case 'n', 'N', ' ':
				var abc payload.PrinterKubeconfigPATH
				abc = printer{ClusterName: cargo.ClusterName, Region: cargo.Region}
				abc.Printer(0)
			}

			break
		}
		fmt.Printf("Waiting.. Status: %v\n", clusterDS.Status)
		time.Sleep(15 * time.Second)
	}

	return nil
}

// kubeconfigDeleter deletes all configs related to the provided cluster
func kubeconfigDeleter(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

// deleteClusterWithID delete cluster from CIVO by provided regionCode and clusterID
func deleteClusterWithID(clusterID, regionCode string) error {
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
	workingDir := kubeconfig + name + "-" + region

	data, err := os.ReadFile(workingDir + "/info")
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	arr := strings.Split(string(data), " ")

	if err = deleteClusterWithID(arr[0], arr[1]); err != nil {
		return err
	}

	if err := kubeconfigDeleter(workingDir); err != nil {
		return err
	}

	var abc payload.PrinterKubeconfigPATH
	abc = printer{ClusterName: name, Region: region}
	abc.Printer(1)
	return nil
}
