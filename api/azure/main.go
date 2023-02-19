/*
Kubesimplify
@author: Dipankar Das
*/

package azure

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	util "github.com/kubesimplify/ksctl/api/utils"
)

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
	generateResourceName(obj)
	if obj.HACluster {
		// TODO: HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
		err := haCreateClusterHandler(ctx, obj)
		if err != nil {
			return err
		}
		log.Printf("Created the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	} else {
		_, err := managedCreateClusterHandler(ctx, obj)
		if err != nil {
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
	if obj.HACluster {
		// TODO: HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
		err := haDeleteClusterHandler(ctx, obj)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "azure", "ha", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region)); err != nil {
			return err
		}
		log.Printf("Deleted the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	} else {

		err := managedDeleteClusterHandler(ctx, obj)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "azure", "managed", obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region)); err != nil {
			return err
		}
		log.Printf("Deleted the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.Config.ResourceGroupName, obj.Region)
	}
	return nil
}

func isPresent(kind string, obj AzureProvider) bool {
	path := util.GetPath(util.CLUSTER_PATH, "azure", kind, obj.ClusterName+" "+obj.Config.ResourceGroupName+" "+obj.Region, "info.json")
	_, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (provider AzureProvider) SwitchContext() error {
	switch provider.HACluster {
	case true:
		if isPresent("ha", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(true, 0)
			return nil
		}
	case false:
		if isPresent("managed", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.Config.ResourceGroupName}
			printKubeconfig.Printer(false, 0)
			return nil
		}
	}
	return fmt.Errorf("ERR Cluster not found")
}
