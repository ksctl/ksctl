package ha_civo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/src/api/payload"
)

// all the configs are present in .ksctl
// want to save the config to ~/.ksctl/config/ha-civo/<Cluster Name> <Region>/*

// TODO: getKubeconfig() fix the path
func getKubeconfig(params ...string) string {
	var ret string

	if runtime.GOOS == "windows" {
		ret = fmt.Sprintf("%s\\.ksctl\\config\\ha-civo", payload.GetUserName())
		for _, item := range params {
			ret += "\\" + item
		}
	} else {
		ret = fmt.Sprintf("%s/.ksctl/config/ha-civo", payload.GetUserName())
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

// isPresent Checks whether the cluster to create is already had been created
func isPresent(clusterName, Region string) bool {
	_, err := os.ReadFile(GetPath(1, clusterName+" "+Region, "info", "network"))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// isValidSize checks whether the size passed by user is valid according to CIVO
func isValidSize(size string) bool {
	validSizes := []string{
		"g3.xsmall",
		"g3.small",
		"g3.medium",
		"g3.large",
		"g3.xlarge",
		"g3.2xlarge"}
	for _, valid := range validSizes {
		if strings.Compare(valid, size) == 0 {
			return true
		}
	}
	fmt.Printf("\n\n\033[34;40mAvailable Node sizes:\n- g3.xsmall\n- g3.small\n- g3.medium\n- g3.large\n- g3.xlarge\n- g3.2xlarge\033[0m\n")
	return false
}

func cleanup(name, region string) error {
	log.Println("[ERR] Cannot continue 😢")
	return DeleteCluster(name, region)
}

// TODO: If error occurs then remove all the resources created
// CreateVM TODO: Handle the errors
// check if the same cluster is present or not
// get the format of cluster name verified and the region as well as the nodeSize
func CreateCluster(name, region, nodeSize string, noCP, noWP int) error {

	if isPresent(name, region) {
		return fmt.Errorf("[FATAL] CLUSTER ALREADY PRESENT")
	}

	if !payload.IsValidRegionCIVO(region) {
		return fmt.Errorf("[INVALID] REGION")
	}

	if !payload.IsValidName(name) {
		return fmt.Errorf("[INVALID] NAME FORMAT")
	}

	if !isValidSize(nodeSize) {
		return fmt.Errorf("INVALID] SIZE")
	}

	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}

	// NOTE: Config Loadbalancer require the control planes privateIPs

	mysqlEndpoint, err := CreateDatabase(client, name)
	if err != nil {
		_ = cleanup(name, region)
		return err
	}

	loadBalancer, err := CreateLoadbalancer(client, name)
	if err != nil {
		_ = cleanup(name, region)
		return err
	}

	var controlPlanes = make([](*civogo.Instance), noCP)

	for i := 0; i < noCP; i++ {
		controlPlanes[i], err = CreateControlPlane(client, i+1, name, nodeSize)
		if err != nil {
			_ = cleanup(name, region)
			return err
		}
	}

	// NOTE: Config the loadbalancer before controlplane is configured

	var controlPlaneIPs = make([]string, noCP)
	for i := 0; i < noCP; i++ {
		controlPlaneIPs[i] = controlPlanes[i].PrivateIP + ":6443"
	}

	err = ConfigLoadBalancer(loadBalancer, controlPlaneIPs)
	if err != nil {
		_ = cleanup(name, region)
		return err
	}

	token := ""
	for i := 0; i < noCP; i++ {
		if i == 0 {
			err = ExecWithoutOutput(controlPlanes[i].PublicIP, controlPlanes[i].InitialPassword, scriptWithoutCP_1(mysqlEndpoint, loadBalancer.PrivateIP), true)
			if err != nil {
				_ = cleanup(name, region)
				return err
			}
			token = GetTokenFromCP_1(controlPlanes[0])
			if len(token) == 0 {
				fmt.Println("Cannot retrieve k3s token")
			}
		} else {
			err = ExecWithoutOutput(controlPlanes[i].PublicIP, controlPlanes[i].InitialPassword, scriptCP_n(mysqlEndpoint, loadBalancer.PrivateIP, token), true)
			if err != nil {
				_ = cleanup(name, region)
				return err
			}
		}
		log.Printf("[CONFIGURED] control-plane-%d\n", i+1)
	}

	kubeconfig, err := FetchKUBECONFIG(controlPlanes[0])
	if err != nil {
		return fmt.Errorf("Cannot fetch kubeconfig\n" + err.Error())
	}
	newKubeconfig := strings.Replace(kubeconfig, "127.0.0.1", loadBalancer.PublicIP, 1)
	fmt.Println(newKubeconfig)

	_ = SaveKubeconfig(name, region, newKubeconfig)

	log.Println("JOINING WORKER NODES")
	var workerPlanes = make([](*civogo.Instance), noWP)

	for i := 0; i < noWP; i++ {
		workerPlanes[i], err = CreateWorkerNode(client, i+1, name, loadBalancer.PrivateIP, token, nodeSize)
		if err != nil {
			_ = cleanup(name, region)
			return err
		}
	}

	log.Println("Created the k3s ha cluster!!")
	return nil
}

func DeleteCluster(name, region string) error {
	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}

	if err := DeleteInstances(client, name); err != nil && !errors.Is(civogo.DatabaseInstanceNotFoundError, err) {
		return err
	}
	time.Sleep(20 * time.Second)

	if err := DeleteFirewalls(client, name); err != nil && !errors.Is(civogo.DatabaseFirewallNotFoundError, err) {
		return err
	}

	time.Sleep(15 * time.Second)
	if err := DeleteNetworks(client, name); err != nil && !errors.Is(civogo.DatabaseNetworkNotFoundError, err) {
		return err
	}

	if err := DeleteAllPaths(name, region); err != nil {
		return err
	}
	return nil
}
