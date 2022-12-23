package ha_civo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/src/api/payload"
)

// all the configs are present in .ksctl
// want to save the config to ~/.ksctl/config/ha-civo/<Cluster Name> <Region>/*

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	file, err := os.ReadFile(payload.GetPathCIVO(0))
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
	_, err := os.ReadFile(payload.GetPathCIVO(1, "ha-civo", clusterName+" "+Region, "info.json"))
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

// cleanup called when error is encountered during creation og cluster
func cleanup(name, region string) error {
	log.Println("[ERR] Cannot continue 😢")
	return DeleteCluster(name, region, false)
}

func validationOfArguments(name, region, nodeSize string) error {
	if isPresent(name, region) {
		return fmt.Errorf("🚨 💀 CLUSTER ALREADY PRESENT")
	}

	if !payload.IsValidRegionCIVO(region) {
		return fmt.Errorf("🚩 REGION")
	}

	if !payload.IsValidName(name) {
		return fmt.Errorf("🚩 NAME FORMAT")
	}

	if !isValidSize(nodeSize) {
		return fmt.Errorf("🚩 SIZE")
	}

	return nil
}

func AddMoreWorkerNodes(name, region, nodeSize string, noWP int) error {

	if !isPresent(name, region) {
		return fmt.Errorf("🚨 💀 CLUSTER NOT PRESENT")
	}
	if !payload.IsValidRegionCIVO(region) {
		return fmt.Errorf("🚩 REGION")
	}

	if !payload.IsValidName(name) {
		return fmt.Errorf("🚩 NAME FORMAT")
	}

	if !isValidSize(nodeSize) {
		return fmt.Errorf("🚩 SIZE")
	}

	config, err := GetConfig(name, region)
	if err != nil {
		return err
	}

	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return err
	}
	var obj HACollection

	obj = &HAType{
		Client:        client,
		NodeSize:      nodeSize,
		ClusterName:   name,
		DiskImgID:     diskImg.ID,
		WPFirewallID:  config.NetworkIDs.FirewallIDWorkerNode,
		NetworkID:     config.NetworkIDs.NetworkID,
		Configuration: &config}

	log.Println("JOINING Additional WORKER NODES")
	lb, err := obj.GetInstance(config.InstanceIDs.LoadBalancerNode[0])
	var workerPlanes = make([](*civogo.Instance), noWP)

	noOfWorkerNodes := len(config.InstanceIDs.WorkerNodes)

	for i := 0; i < noWP; i++ {
		workerPlanes[i], err = obj.CreateWorkerNode(i+noOfWorkerNodes+1, lb.PrivateIP, config.ServerToken)
		if err != nil {
			log.Println("Failed to add more nodes..")
			return err
		}
	}

	log.Printf("\n🗒 Currently no firewall Rules are being used so you can add them using CIVO Dashboard\n")
	log.Println("Added more nodes 🥳 🎉 ")
	return nil
}

func DeleteSomeWorkerNodes(clusterName, region string, noWP int) error {
	if !payload.IsValidRegionCIVO(region) {
		return fmt.Errorf("🚩 REGION")
	}

	if !payload.IsValidName(clusterName) {
		return fmt.Errorf("🚩 NAME FORMAT")
	}

	if !isPresent(clusterName, region) {
		return fmt.Errorf("🚨 💀 CLUSTER NOT PRESENT")
	}

	log.Printf(`NOTE 🚨
((Deleteion of nodes happens from most recent added to first created worker node))
i.e. of workernodes 1, 2, 3, 4
then deletion will happen from 4, 3, 2, 1
1) make sure you first drain the no of nodes
		kubectl drain node <node name>
2) then delete before deleting the instance
		kubectl delete node <node name>
`)
	fmt.Println("Enter your choice to continue..[y/N]")
	choice := "n"
	unsafe := false
	fmt.Scanf("%s", &choice)
	if strings.Compare("y", choice) == 0 ||
		strings.Compare("yes", choice) == 0 ||
		strings.Compare("Y", choice) == 0 {
		unsafe = true
	}

	if !unsafe {
		return nil
	}

	config, err := GetConfig(clusterName, region)
	if err != nil {
		return err
	}

	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}

	var obj HACollection

	obj = &HAType{
		Client:        client,
		NodeSize:      "",
		ClusterName:   clusterName,
		DiskImgID:     "",
		WPFirewallID:  config.NetworkIDs.FirewallIDWorkerNode,
		NetworkID:     config.NetworkIDs.NetworkID,
		Configuration: &config}

	currNoOfWorkerNodes := len(config.InstanceIDs.WorkerNodes)
	if noWP > currNoOfWorkerNodes {
		return fmt.Errorf("Requested no of deletion is more than present")
	}

	for i := 0; i < noWP; i++ {
		err := obj.DeleteInstance(config.InstanceIDs.WorkerNodes[len(config.InstanceIDs.WorkerNodes)-1])
		if err != nil {
			return err
		}

		config.InstanceIDs.WorkerNodes = config.InstanceIDs.WorkerNodes[:len(config.InstanceIDs.WorkerNodes)-1]

		err = saveConfig(clusterName+" "+region, config)
		if err != nil {
			return err
		}
	}

	log.Println("Deleted some nodes 🥳 🎉 ")
	return nil
}

// this will be made available to each create functional calls
func CreateCluster(name, region, nodeSize string, noCP, noWP int) error {

	if errV := validationOfArguments(name, region, nodeSize); errV != nil {
		return errV
	}

	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return err
	}

	var obj HACollection

	obj = &HAType{
		Client:       client,
		NodeSize:     nodeSize,
		ClusterName:  name,
		DiskImgID:    diskImg.ID,
		DBFirewallID: "",
		LBFirewallID: "",
		CPFirewallID: "",
		WPFirewallID: "",
		NetworkID:    "",
		Configuration: &JsonStore{
			ClusterName: name,
			Region:      region,
			DBEndpoint:  "",
			ServerToken: "",
			InstanceIDs: InstanceID{},
			NetworkIDs:  NetworkID{},
		},
	}

	// NOTE: Config Loadbalancer require the control planes privateIPs

	mysqlEndpoint, err := obj.CreateDatabase()
	if err != nil {
		_ = cleanup(name, region)
		return err
	}

	loadBalancer, err := obj.CreateLoadbalancer()
	if err != nil {
		_ = cleanup(name, region)
		return err
	}

	var controlPlanes = make([](*civogo.Instance), noCP)

	for i := 0; i < noCP; i++ {
		controlPlanes[i], err = obj.CreateControlPlane(i + 1)
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
			token = obj.GetTokenFromCP_1(controlPlanes[0])
			if len(token) == 0 {
				_ = cleanup(name, region)
				return fmt.Errorf("🚨 Cannot retrieve k3s token")
			}
		} else {
			err = ExecWithoutOutput(controlPlanes[i].PublicIP, controlPlanes[i].InitialPassword, scriptCP_n(mysqlEndpoint, loadBalancer.PrivateIP, token), true)
			if err != nil {
				_ = cleanup(name, region)
				return err
			}
		}
		log.Printf("✅ 🔧 control-plane-%d\n", i+1)
	}

	kubeconfig, err := FetchKUBECONFIG(controlPlanes[0])
	if err != nil {
		return fmt.Errorf("Cannot fetch kubeconfig\n" + err.Error())
	}
	newKubeconfig := strings.Replace(kubeconfig, "127.0.0.1", loadBalancer.PublicIP, 1)
	fmt.Println(newKubeconfig)

	_ = obj.SaveKubeconfig(newKubeconfig)

	log.Println("JOINING WORKER NODES")
	var workerPlanes = make([](*civogo.Instance), noWP)

	for i := 0; i < noWP; i++ {
		workerPlanes[i], err = obj.CreateWorkerNode(i+1, loadBalancer.PrivateIP, token)
		if err != nil {
			_ = cleanup(name, region)
			return err
		}
	}

	log.Println("Created the k3s ha cluster!!🥳 🎉 ")

	var printKubeconfig payload.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(0)
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
		fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'\n", payload.GetPathCIVO(1, "ha-civo", p.ClusterName+" "+p.Region, "config")))
	case 1:
		fmt.Printf("\n\033[33;40mUse the following command to unset KUBECONFIG\033[0m\n\n")
		fmt.Println(fmt.Sprintf("unset KUBECONFIG\n"))
	}
	fmt.Println()
}

// DeleteCluster to delete the entire cluster
func DeleteCluster(name, region string, showMsg bool) error {
	client, err := civogo.NewClient(fetchAPIKey(), region)
	if err != nil {
		return err
	}
	var obj HACollection

	obj = &HAType{
		Client:       client,
		NodeSize:     "",
		ClusterName:  name,
		DiskImgID:    "",
		DBFirewallID: "",
		LBFirewallID: "",
		CPFirewallID: "",
		WPFirewallID: "",
		NetworkID:    ""}

	if showMsg {
		log.Printf(`NOTE 🚨
THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'\n`, name+" "+region)
		fmt.Println("Enter your choice to continue..[y/N]")
		choice := "n"
		unsafe := false
		fmt.Scanf("%s", &choice)
		if strings.Compare("y", choice) == 0 ||
			strings.Compare("yes", choice) == 0 ||
			strings.Compare("Y", choice) == 0 {
			unsafe = true
		}

		if !unsafe {
			return nil
		}
	}

	if err := obj.DeleteInstances(); err != nil && !errors.Is(civogo.DatabaseInstanceNotFoundError, err) {
		return err
	}
	time.Sleep(10 * time.Second)

	if err := obj.DeleteNetworks(); err != nil && !errors.Is(civogo.DatabaseNetworkNotFoundError, err) {
		return err
	}

	if err := DeleteAllPaths(name, region); err != nil {
		return err
	}

	log.Println("Deleted the cluster 🏭 🔥")

	var printKubeconfig payload.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(1)
	return nil
}
