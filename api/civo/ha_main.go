package civo

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/civo/civogo"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func isValidSizeHA(size string) bool {
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

func haCreateClusterHandler(name, region, nodeSize string, noCP, noWP int) error {

	if errV := validationOfArguments(name, region); errV != nil {
		return errV
	}

	if !isValidSizeHA(nodeSize) {
		return fmt.Errorf("🚩 SIZE")
	}

	if isPresent("ha", name, region) {
		return fmt.Errorf("🚨 💀 CLUSTER ALREADY PRESENT")
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
		SSHID:        "",
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

	err = obj.UploadSSHKey()
	if err != nil {
		return err
	}

	mysqlEndpoint, err := obj.CreateDatabase()
	if err != nil {
		return err
	}

	loadBalancer, err := obj.CreateLoadbalancer()
	if err != nil {
		return err
	}

	var controlPlanes = make([](*civogo.Instance), noCP)

	for i := 0; i < noCP; i++ {
		controlPlanes[i], err = obj.CreateControlPlane(i + 1)
		if err != nil {
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
		return err
	}

	token := ""
	for i := 0; i < noCP; i++ {
		if i == 0 {
			err = ExecWithoutOutput(controlPlanes[i].PublicIP, controlPlanes[i].InitialPassword, scriptWithoutCP_1(mysqlEndpoint, loadBalancer.PrivateIP), true)
			if err != nil {
				return err
			}
			token = obj.GetTokenFromCP_1(controlPlanes[0])
			if len(token) == 0 {
				return fmt.Errorf("🚨 Cannot retrieve k3s token")
			}
		} else {
			err = ExecWithoutOutput(controlPlanes[i].PublicIP, controlPlanes[i].InitialPassword, scriptCP_n(mysqlEndpoint, loadBalancer.PrivateIP, token), true)
			if err != nil {
				return err
			}
		}
		log.Printf("✅ Configured control-plane-%d\n", i+1)
	}

	kubeconfig, err := FetchKUBECONFIG(controlPlanes[0])
	if err != nil {
		return fmt.Errorf("Cannot fetch kubeconfig\n" + err.Error())
	}
	newKubeconfig := strings.Replace(kubeconfig, "127.0.0.1", loadBalancer.PublicIP, 1)

	newKubeconfig = strings.Replace(newKubeconfig, "default", name+"-"+strings.ToLower(region)+"-ha-civo", -1)

	err = obj.SaveKubeconfig(newKubeconfig)
	if err != nil {
		return err
	}

	log.Println("⛓  JOINING WORKER NODES")
	var workerPlanes = make([](*civogo.Instance), noWP)

	for i := 0; i < noWP; i++ {
		workerPlanes[i], err = obj.CreateWorkerNode(i+1, loadBalancer.PrivateIP, token)
		if err != nil {
			return err
		}
	}

	log.Println("Created your HA Civo cluster!!🥳 🎉 ")
	log.Printf("\n🗒 Currently no firewall Rules are being used so you can add them using CIVO Dashboard\n")
	fmt.Println(`
NOTE
for the very first kubectl API call, do this
	kubectl cluster-info --insecure-skip-tls-verify

after this you can procede with normal oprtation of the cluster`)

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(true, 0)
	return nil
}

func haDeleteClusterHandler(name, region string, showMsg bool) error {

	if errV := validationOfArguments(name, region); errV != nil {
		return errV
	}

	if !isPresent("ha", name, region) {
		return fmt.Errorf("🚨 💀 CLUSTER NOT PRESENT")
	}

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
THIS IS A DESTRUCTIVE STEP MAKE SURE IF YOU WANT TO DELETE THE CLUSTER '%s'
`, name+" "+region)
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

	var errR error

	if err := obj.DeleteInstances(); err != nil && !errors.Is(civogo.DatabaseInstanceNotFoundError, err) {
		return err
	}
	errR = err
	if errR != nil {
		// dont delete the configs
		return errR
	}
	time.Sleep(10 * time.Second)

	errR = nil
	if err := obj.DeleteNetworks(); err != nil && !errors.Is(civogo.DatabaseNetworkNotFoundError, err) {
		return err
	}
	errR = err
	if errR != nil {
		// dont delete the configs
		return errR
	}

	err = obj.DeleteSSHKeyPair()
	if err != nil {
		return err
	}

	if err := DeleteAllPaths(name, region); err != nil {
		return err
	}

	log.Println("Deleted the cluster 🏭 🔥")

	var printKubeconfig util.PrinterKubeconfigPATH
	printKubeconfig = printer{ClusterName: name, Region: region}
	printKubeconfig.Printer(true, 1)
	return nil
}

func (provider CivoProvider) AddMoreWorkerNodes() error {
	name, region, nodeSize, noWP := provider.ClusterName, provider.Region, provider.Spec.Disk, provider.Spec.HAWorkerNodes

	if errV := validationOfArguments(name, region); errV != nil {
		return errV
	}

	if !isValidSizeHA(nodeSize) {
		return fmt.Errorf("🚩 SIZE")
	}

	if !isPresent("ha", name, region) {
		return fmt.Errorf("🚨 💀 CLUSTER NOT PRESENT")
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

	log.Println("Added more nodes 🥳 🎉 ")
	return nil
}

func (provider CivoProvider) DeleteSomeWorkerNodes() error {
	clusterName := provider.ClusterName
	region := provider.Region
	noWP := provider.Spec.HAWorkerNodes
	if !util.IsValidRegionCIVO(region) {
		return fmt.Errorf("🚩 REGION")
	}

	if !util.IsValidName(clusterName) {
		return fmt.Errorf("🚩 NAME FORMAT")
	}

	if !isPresent("ha", clusterName, region) {
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
