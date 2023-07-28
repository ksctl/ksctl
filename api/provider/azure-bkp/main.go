/*
Kubesimplify
@author: Dipankar Das
*/

package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	log "github.com/kubesimplify/ksctl/api/provider/logger"
	util "github.com/kubesimplify/ksctl/api/provider/utils"
)

// TODO: add the VMSize as user defined option

func Credentials(logger log.Logger) bool {
	logger.Print("Enter your SUBSCRIPTION ID 👇")
	skey, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	logger.Print("Enter your TENANT ID 👇")
	tid, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	logger.Print("Enter your CLIENT ID 👇")
	cid, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	logger.Print("Enter your CLIENT SECRET 👇")
	cs, err := util.UserInputCredentials(logger)
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.AzureCredential{
		SubscriptionID: skey,
		TenantID:       tid,
		ClientID:       cid,
		ClientSecret:   cs,
	}

	err = util.SaveCred(logger, apiStore, "azure")

	if err != nil {
		logger.Err(err.Error())
		return false
	}
	return true

}

type AzureProvider struct {
	ClusterName    string                 `json:"cluster_name"`
	HACluster      bool                   `json:"ha_cluster"`
	Region         string                 `json:"region"`
	Spec           util.Machine           `json:"spec"`
	SubscriptionID string                 `json:"subscription_id"`
	Config         *AzureStateCluster     `json:"config"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"`
	SSH_Payload    *util.SSHPayload       `json:"ssh___payload"`
}

// AddMoreWorkerNodes adds more worker nodes to the existing HA cluster
func (obj *AzureProvider) AddMoreWorkerNodes(logging log.Logger) error {

	// logging := log.Logger{Verbose: true} // make it move to cli part
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
	err := setRequiredENV_VAR(logging, ctx, obj)
	if err != nil {
		return err
	}
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

	err = obj.ConfigReader(logging, "ha")
	if err != nil {
		return fmt.Errorf("Unable to read configuration: %v", err)
	}
	obj.AzureTokenCred = cred

	logging.Info("JOINING Additional WORKER NODES", "")

	noOfWorkerNodes := len(obj.Config.InfoWorkerPlanes.Names)

	for i := 0; i < obj.Spec.HAWorkerNodes; i++ {
		err := obj.createWorkerPlane(logging, ctx, i+noOfWorkerNodes+1)
		if err != nil {
			logging.Err("Failed to add more nodes..")
			return err
		}
	}

	logging.Info("Added more nodes 🥳 🎉 ")
	return nil
}

// DeleteSomeWorkerNodes deletes workerNodes from existing HA cluster
func (obj *AzureProvider) DeleteSomeWorkerNodes(logging log.Logger) error {

	// logging := log.Logger{Verbose: true} // make it move to cli part
	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !util.IsValidName(obj.ClusterName) {
		return fmt.Errorf("invalid cluster name: %v", obj.ClusterName)
	}

	if !isValidRegion(obj.Region) {
		return fmt.Errorf("region {%s} is invalid", obj.Region)
	}

	logging.Note(`🚨 ((Deleteion of nodes happens from most recent added to first created worker node))
i.e. of workernodes 1, 2, 3, 4
then deletion will happen from 4, 3, 2, 1
1) make sure you first drain the no of nodes
		kubectl drain node <node name>
2) then delete before deleting the instance
		kubectl delete node <node name>
`)
	logging.Print("Enter your choice to continue..[y/N]")
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
	err := setRequiredENV_VAR(logging, ctx, obj)
	if err != nil {
		return err
	}
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

	err = obj.ConfigReader(logging, "ha")
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
		err := obj.DeleteVM(ctx, logging, obj.Config.InfoWorkerPlanes.Names[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeleteDisk(ctx, logging, obj.Config.InfoWorkerPlanes.DiskNames[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeleteNetworkInterface(ctx, logging, obj.Config.InfoWorkerPlanes.NetworkInterfaceNames[currLen-1])
		if err != nil {
			return err
		}

		err = obj.DeletePublicIP(ctx, logging, obj.Config.InfoWorkerPlanes.PublicIPNames[currLen-1])
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

		err = obj.ConfigWriter(logging, "ha")
		if err != nil {
			return err
		}
	}

	logging.Info("Deleted some nodes 🥳 🎉 ")
	return nil
}

func (obj *AzureProvider) CreateCluster(logging log.Logger) error {

	ctx := context.Background()
	err := setRequiredENV_VAR(logging, ctx, obj)
	if err != nil {
		return err
	}
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

		if isPresent("ha", *obj) {
			return fmt.Errorf("cluster already exists: %v", obj.ClusterName)
		}
		err := haCreateClusterHandler(ctx, logging, obj)
		if err != nil {
			logging.Err("CLEANUP TRIGGERED!: failed to create")
			_ = haDeleteClusterHandler(ctx, logging, obj, false)
			return err
		}
	} else {
		obj.Config.ResourceGroupName = obj.ClusterName + "-ksctl"

		if isPresent("managed", *obj) {
			return fmt.Errorf("cluster already exists: %v", obj.ClusterName)
		}
		_, err := managedCreateClusterHandler(ctx, logging, obj)
		if err != nil {
			logging.Err("CLEANUP TRIGGERED!: failed to create")
			_ = managedDeleteClusterHandler(ctx, logging, obj, false)
			return err
		}
	}
	return nil
}

func (obj *AzureProvider) DeleteCluster(logging log.Logger) error {

	ctx := context.Background()
	err := setRequiredENV_VAR(logging, ctx, obj)
	if err != nil {
		return err
	}
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
		if !isPresent("ha", *obj) {
			return fmt.Errorf("cluster doesn't exists: %v", obj.ClusterName)
		}
		err := haDeleteClusterHandler(ctx, logging, obj, true)
		if err != nil {
			return err
		}

	} else {

		obj.Config.ResourceGroupName = obj.ClusterName + "-ksctl"
		if !isPresent("managed", *obj) {
			return fmt.Errorf("cluster doesn't exists: %v", obj.ClusterName)
		}
		err := managedDeleteClusterHandler(ctx, logging, obj, true)
		if err != nil {
			return err
		}

	}
	return nil
}

func (provider AzureProvider) SwitchContext(logging log.Logger) error {
	provider.Config = &AzureStateCluster{}
	switch provider.HACluster {
	case true:
		provider.Config.ResourceGroupName = provider.ClusterName + "-ha-ksctl"
		if isPresent("ha", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(logging, true, 0)
			return nil
		}
	case false:
		provider.Config.ResourceGroupName = provider.ClusterName + "-ksctl"
		if isPresent("managed", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(logging, false, 0)
			return nil
		}
	}
	return fmt.Errorf("ERR Cluster not found")
}
