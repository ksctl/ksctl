/*
Kubesimplify
Credit to @civo
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package civo

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	util "github.com/kubesimplify/ksctl/api/utils"

	"github.com/civo/civogo"
)

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(util.GetPathCIVO(0))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

func Credentials() bool {
	file, err := os.OpenFile(util.GetPathCIVO(0), os.O_WRONLY, 0640)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	apikey := ""
	func() {
		fmt.Println("Enter your API-TOKEN-KEY: ")
		_, err = fmt.Scan(&apikey)
		if err != nil {
			panic(err.Error())
		}
	}()

	_, err = file.Write([]byte(fmt.Sprintf("API-TOKEN-Key: %s", apikey)))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// kubeconfigWriter Writes kubeconfig supplied to config directory of respective cluster created
func kubeconfigWriter(kubeconfig, clusterN, region, clusterID string) error {
	// create the necessary folders and files
	clusterFolder := clusterN + " " + region
	err := os.Mkdir(util.GetPathCIVO(1, "civo", clusterFolder), 0750)

	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(util.GetPathCIVO(1, "civo", clusterFolder, "config"))
	if err != nil {
		return err
	}
	_, err = os.Create(util.GetPathCIVO(1, "civo", clusterFolder, "info"))
	if err != nil {
		return err
	}

	// write the contents to the req. files
	file, err := os.OpenFile(util.GetPathCIVO(1, "civo", clusterFolder, "config"), os.O_WRONLY, 0750)
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
	if err = os.WriteFile(util.GetPathCIVO(1, "civo", clusterFolder, "info"),
		[]byte(fmt.Sprintf("%s %s", clusterID, region)), 0640); err != nil {
		return err
	}

	//FIXME: make this more reliable ISSUE #5
	err = os.Setenv("KUBECONFIG", util.GetPathCIVO(1, "civo", clusterFolder, "config"))
	if err != nil {
		return err
	}
	return nil
}

// ClusterInfoInjecter Serializes the information which is return as utils.CivoProvider for sending it to API
// {clustername, regionCode, Size of Nodes, No of nodes, Applications(optional), cniPlugion(optional)}
func ClusterInfoInjecter(clusterName, reg, size string, noOfNodes int, application, cniPlugin string) util.CivoProvider {

	if len(application) == 0 {
		application = "Traefik-v2-nodeport,metrics-server" // default: applications
	} else {
		application += ",Traefik-v2-nodeport,metrics-server"
	}
	if len(cniPlugin) == 0 {
		cniPlugin = "flannel" // default: flannel
	}

	spec := util.CivoProvider{
		ClusterName: clusterName,
		Region:      reg,
		APIKey:      fetchAPIKey(),
		HACluster:   false,
		Spec: util.Machine{
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
	_, err := os.ReadFile(util.GetPathCIVO(1, "civo", clusterName+" "+Region, "info"))
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
	fmt.Printf("\n\n\033[34;40mAvailable Node sizes:\n- g4s.kube.xsmall\n- g4s.kube.small\n- g4s.kube.medium\n- g4s.kube.large\n- g4p.kube.small\n- g4p.kube.medium\n- g4p.kube.large\n- g4p.kube.xlarge\n- g4c.kube.small\n- g4c.kube.medium\n- g4c.kube.large\n- g4c.kube.xlarge\n- g4m.kube.small\n- g4m.kube.medium\n- g4m.kube.large\n- g4m.kube.xlarge\033[0m\n")
	return false
}

// CreateCluster creates cluster as provided configuration and returns whether it fails or not
func CreateCluster(civoConfig util.CivoProvider) error {
	if len(civoConfig.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !util.IsValidName(civoConfig.ClusterName) {
		return fmt.Errorf("INVALID CLUSTER NAME")
	}

	if !util.IsValidRegionCIVO(civoConfig.Region) {
		return fmt.Errorf("region code is Invalid")
	}

	if isPresent(civoConfig.ClusterName, civoConfig.Region) {
		return fmt.Errorf("DUPLICATE Cluster")
	}

	if !isValidSize(civoConfig.Spec.Disk) {
		return fmt.Errorf("INVALID size of node")
	}

	client, err := civogo.NewClient(civoConfig.APIKey, civoConfig.Region)
	if err != nil {
		return err
	}

	defaultNetwork, err := client.GetDefaultNetwork()
	if err != nil {
		return err
	}

	configK8s := &civogo.KubernetesClusterConfig{
		Name:            civoConfig.ClusterName,
		Region:          civoConfig.Region,
		NumTargetNodes:  civoConfig.Spec.Nodes,
		TargetNodesSize: civoConfig.Spec.Disk,
		NetworkID:       defaultNetwork.ID,
		Applications:    civoConfig.Application,
		CNIPlugin:       civoConfig.CNIPlugin,
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
	for {
		// clusterDS fetches the current state of kubernetes cluster given its id
		clusterDS, _ := client.GetKubernetesCluster(resp.ID)
		if clusterDS.Ready {

			err := kubeconfigWriter(clusterDS.KubeConfig, civoConfig.ClusterName, civoConfig.Region, resp.ID)
			if err != nil {
				return err
			}

			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: civoConfig.ClusterName, Region: civoConfig.Region}
			printKubeconfig.Printer(0)

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

	// data will contain the saved ClusterID and Region
	data, err := os.ReadFile(util.GetPathCIVO(1, "civo", name+" "+region, "info"))
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	arr := strings.Split(string(data), " ")

	if err = deleteClusterWithID(arr[0], arr[1]); err != nil {
		return err
	}

	if err := kubeconfigDeleter(util.GetPathCIVO(1, "civo", name+" "+region)); err != nil {
		return err
	}

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(1)
	return nil
}

type printer struct {
	ClusterName string
	Region      string
}

func (p printer) Printer(a int) {
	switch a {
	case 0:
		fmt.Printf("\n\033[33;40mTo use this cluster set this environment variable\033[0m\n\n")
		fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'", util.GetPathCIVO(1, "civo", p.ClusterName+" "+p.Region, "config")))
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG"))
	}
	fmt.Println()
}

// SwitchContext TODO: Add description
func SwitchContext(clusterName, region string) error {
	if isPresent(clusterName, region) {
		// TODO: ISSUE #5
		var printKubeconfig util.PrinterKubeconfigPATH
		printKubeconfig = printer{ClusterName: clusterName, Region: region}
		printKubeconfig.Printer(0)
		return nil
	}
	return fmt.Errorf("ERR Cluster not found")
}
