/*
Kubesimplify
@author: Dipankar Das
*/

package azure

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	util "github.com/kubesimplify/ksctl/api/utils"
)

// TODO: add the VMSize as user defined option

func Credentials() bool {
	fmt.Println("Enter your SUBSCRIPTION ID 👇")
	skey, err := util.UserInputCredentials()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your TENANT ID 👇")
	tid, err := util.UserInputCredentials()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your CLIENT ID 👇")
	cid, err := util.UserInputCredentials()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your CLIENT SECRET 👇")
	cs, err := util.UserInputCredentials()
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.AzureCredential{
		SubscriptionID: skey,
		TenantID:       tid,
		ClientID:       cid,
		ClientSecret:   cs,
	}

	err = util.SaveCred(apiStore, "azure")

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true

}

type AzureProvider struct {
	ClusterName    string
	HACluster      bool
	Region         string
	Spec           util.Machine
	SubscriptionID string
	Config         *AzureStateCluster
	AzureTokenCred azcore.TokenCredential
	SSH_Payload    *util.SSHPayload
}

// AddMoreWorkerNodes adds more worker nodes to the existing HA cluster
func (obj *AzureProvider) AddMoreWorkerNodes() error {

	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}
	if !isValidNodeSize(obj.Spec.Disk) {
		return fmt.Errorf("node size {%s} is invalid", obj.Spec.Disk)
	}

	if !isValidRegion(obj.Region) {
		return fmt.Errorf("region {%s} is invalid", obj.Region)
	}

	ctx := context.Background()
	setRequiredENV_VAR(ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	obj.Config.ResourceGroupName = obj.ClusterName + "-ha-ksctl"
	if !isPresent("ha", *obj) {
		return fmt.Errorf("cluster does not exists: %v", obj.ClusterName)
	}

	err = obj.ConfigReader("ha")
	if err != nil {
		return fmt.Errorf("Unable to read configuration: %v", err)
	}
	obj.AzureTokenCred = cred

	log.Println("JOINING Additional WORKER NODES")

	noOfWorkerNodes := len(obj.Config.InfoWorkerPlanes.Names)

	for i := 0; i < obj.Spec.HAWorkerNodes; i++ {
		err := obj.createWorkerPlane(ctx, i+noOfWorkerNodes+1)
		if err != nil {
			log.Fatalf("Failed to add more nodes..")
		}
	}

	log.Println("Added more nodes 🥳 🎉 ")
	return nil
}

// DeleteSomeWorkerNodes deletes workerNodes from existing HA cluster
func (obj *AzureProvider) DeleteSomeWorkerNodes() error {

	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !isValidRegion(obj.Region) {
		return fmt.Errorf("region {%s} is invalid", obj.Region)
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

	ctx := context.Background()
	setRequiredENV_VAR(ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	obj.Config.ResourceGroupName = obj.ClusterName + "-ha-ksctl"
	if !isPresent("ha", *obj) {
		return fmt.Errorf("cluster does not exists: %v", obj.ClusterName)
	}

	err = obj.ConfigReader("ha")
	if err != nil {
		return fmt.Errorf("Unable to read configuration: %v", err)
	}
	obj.AzureTokenCred = cred

	requestedNoOfWP := obj.Spec.HAWorkerNodes

	currNoOfWorkerNodes := len(obj.Config.InfoWorkerPlanes.Names)
	if requestedNoOfWP > currNoOfWorkerNodes {
		return fmt.Errorf("Requested no of deletion is more than present")
	}

	for i := 0; i < requestedNoOfWP; i++ {

		currLen := len(obj.Config.InfoWorkerPlanes.Names)
		err := obj.DeleteVM(ctx, obj.Config.InfoWorkerPlanes.Names[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeleteDisk(ctx, obj.Config.InfoWorkerPlanes.DiskNames[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeleteNetworkInterface(ctx, obj.Config.InfoWorkerPlanes.NetworkInterfaceNames[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeletePublicIP(ctx, obj.Config.InfoWorkerPlanes.PublicIPNames[currLen-1])
		if err != nil {
			return err
		}

		// In anyone wants to make a seperate function to update the credentials for workerplane
		obj.Config.InfoWorkerPlanes.Names = obj.Config.InfoWorkerPlanes.Names[:currLen-1]
		obj.Config.InfoWorkerPlanes.DiskNames = obj.Config.InfoWorkerPlanes.DiskNames[:currLen-1]
		obj.Config.InfoWorkerPlanes.NetworkInterfaceNames = obj.Config.InfoWorkerPlanes.NetworkInterfaceNames[:currLen-1]
		obj.Config.InfoWorkerPlanes.PublicIPNames = obj.Config.InfoWorkerPlanes.PublicIPNames[:currLen-1]
		obj.Config.InfoWorkerPlanes.PrivateIPs = obj.Config.InfoWorkerPlanes.PublicIPs[:currLen-1]
		obj.Config.InfoWorkerPlanes.PublicIPs = obj.Config.InfoWorkerPlanes.PublicIPs[:currLen-1]

		err = obj.ConfigWriter("ha")
		if err != nil {
			return err
		}
	}

	log.Println("Deleted some nodes 🥳 🎉 ")
	return nil
}

func (obj *AzureProvider) CreateCluster() error {

	ctx := context.Background()
	setRequiredENV_VAR(ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	if obj.HACluster {
		obj.Config.ResourceGroupName = obj.ClusterName + "-ha-ksctl"

		err := haCreateClusterHandler(ctx, obj)
		if err != nil {
			log.Println("CLEANUP TRIGGERED!: failed to create")
			_ = haDeleteClusterHandler(ctx, obj, false)
			return err
		}
		log.Printf("Created the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	} else {
		obj.Config.ResourceGroupName = obj.ClusterName + "-ksctl"

		_, err := managedCreateClusterHandler(ctx, obj)
		if err != nil {
			_ = managedDeleteClusterHandler(ctx, obj, false)
			return err
		}
		log.Printf("Created the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	}
	return nil
}

func (obj *AzureProvider) DeleteCluster() error {
	ctx := context.Background()
	setRequiredENV_VAR(ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	if obj.HACluster {
		obj.Config.ResourceGroupName = obj.ClusterName + "-ha-ksctl"
		err := haDeleteClusterHandler(ctx, obj, true)
		if err != nil {
			return err
		}

		log.Printf("Deleted the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	} else {
		obj.Config.ResourceGroupName = obj.ClusterName + "-ksctl"
		err := managedDeleteClusterHandler(ctx, obj, true)
		if err != nil {
			return err
		}

		log.Printf("Deleted the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	}
	return nil
}

func (provider AzureProvider) SwitchContext() error {
	provider.Config = &AzureStateCluster{}
	switch provider.HACluster {
	case true:
		provider.Config.ResourceGroupName = provider.ClusterName + "-ha-ksctl"
		if isPresent("ha", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(true, 0)
			return nil
		}
	case false:
		provider.Config.ResourceGroupName = provider.ClusterName + "-ksctl"
		if isPresent("managed", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(false, 0)
			return nil
		}
	}
	return fmt.Errorf("ERR Cluster not found")
}
