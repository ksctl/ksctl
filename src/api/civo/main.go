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
	"runtime"
	"strings"
	"time"

	"github.com/kubesimplify/ksctl/src/api/payload"

	"github.com/civo/civogo"
)

func getKubeconfig(params ...string) string {
	var ret string

	if runtime.GOOS == "windows" {
		ret = fmt.Sprintf("%s\\.ksctl\\config\\civo", payload.GetUserName())
		for _, item := range params {
			ret += "\\" + item
		}
	} else {
		ret = fmt.Sprintf("%s/.ksctl/config/civo", payload.GetUserName())
		for _, item := range params {
			ret += "/" + item
		}
	}
	return ret
}

func getCredentials() string {

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\civo", payload.GetUserName())
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/civo", payload.GetUserName())
	}
}

// GetPath use this in every function and differentiate the logic by using if-else
// flag is used to indicate 1 -> KUBECONFIG, 0 -> CREDENTIALS
func GetPath(flag int8, params ...string) string {
	switch flag {
	case 1:
		return getKubeconfig(params...)
	case 0:
		return getCredentials()
	default:
		return ""
	}
}

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(GetPath(0))
	if err != nil {
		return ""
	}
	if len(file) == 0 {
		return ""
	}

	return strings.Split(string(file), " ")[1]
}

func Credentials() bool {
	file, err := os.OpenFile(GetPath(0), os.O_WRONLY, 0640)
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
	//workingDir := KUBECONFIG_PATH + clusterN + " " + region
	clusterFolder := clusterN + " " + region
	//err := os.Mkdir(workingDir, 0750)
	err := os.Mkdir(GetPath(1, clusterFolder), 0750)

	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = os.Create(GetPath(1, clusterFolder, "config"))
	if err != nil {
		return err
	}
	_, err = os.Create(GetPath(1, clusterFolder, "info"))
	if err != nil {
		return err
	}

	// write the contents to the req. files
	file, err := os.OpenFile(GetPath(1, clusterFolder, "config"), os.O_WRONLY, 0750)
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
	if err = os.WriteFile(GetPath(1, clusterFolder, "info"),
		[]byte(fmt.Sprintf("%s %s", clusterID, region)), 0640); err != nil {
		return err
	}

	//FIXME: make this more reliable ISSUE #5
	err = os.Setenv("KUBECONFIG", GetPath(1, clusterFolder, "config"))
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
	_, err := os.ReadFile(GetPath(1, clusterName+" "+Region, "info"))
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
func CreateCluster(cargo payload.CivoProvider) error {
	if len(cargo.APIKey) == 0 {
		return fmt.Errorf("CREDENTIALS NOT PRESENT")
	}

	if !payload.IsValidName(cargo.ClusterName) {
		return fmt.Errorf("INVALID CLUSTER NAME")
	}

	if !payload.IsValidRegionCIVO(cargo.Region) {
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
			default:
				var printKubeconfig payload.PrinterKubeconfigPATH
				printKubeconfig = printer{ClusterName: cargo.ClusterName, Region: cargo.Region}
				printKubeconfig.Printer(0)
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
	//workingDir := KUBECONFIG_PATH + name + " " + region

	// data will contain the saved ClusterID and Region
	data, err := os.ReadFile(GetPath(1, name+" "+region, "info"))
	if err != nil {
		return fmt.Errorf("NO matching cluster found")
	}

	arr := strings.Split(string(data), " ")

	if err = deleteClusterWithID(arr[0], arr[1]); err != nil {
		return err
	}

	if err := kubeconfigDeleter(GetPath(1, name+" "+region)); err != nil {
		return err
	}

	var printKubeconfig payload.PrinterKubeconfigPATH
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
		fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'", GetPath(1, p.ClusterName+" "+p.Region, "config")))
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
		var printKubeconfig payload.PrinterKubeconfigPATH
		printKubeconfig = printer{ClusterName: clusterName, Region: region}
		printKubeconfig.Printer(0)
		return nil
	}
	return fmt.Errorf("ERR Cluster not found")
}
