/*
Kubesimplify
@maintainer:
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

// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(util.CREDENTIAL_PATH, "azure"))
	if err != nil {
		return ""
	}
	return ""
}

func Credentials() bool {
	skey := ""
	tid := ""
	pi := ""
	pk := ""
	fmt.Println("Enter your SUBSCRIPTION ID: ")
	_, err := fmt.Scan(&skey)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your TENANT ID: ")
	_, err = fmt.Scan(&tid)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your CLIENT ID: ")
	_, err = fmt.Scan(&pi)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter your CLIENT SECRET: ")
	_, err = fmt.Scan(&pk)
	if err != nil {
		panic(err.Error())
	}

	apiStore := util.AzureCredential{
		SubscriptionID: skey,
		TenantID:       tid,
		ClientID:       pi,
		ClientSecret:   pk,
	}

	err = util.SaveCred(apiStore, "azure")

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true

}

type AzureProvider struct {
	ClusterName       string
	HACluster         bool
	Region            string
	Spec              util.Machine
	SubscriptionID    string
	Config            *AzureManagedState
	ResourceGroupName string
	AzureTokenCred    azcore.TokenCredential
}

func (obj *AzureProvider) CreateCluster() error {

	ctx := context.Background()
	setRequiredENV_VAR(ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	generateResourceName(obj)

	if obj.HACluster {
		// HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
	} else {
		obj.AzureTokenCred = cred
		obj.Config = &AzureManagedState{}
		_, err := managedCreateClusterHandler(ctx, obj)
		if err != nil {
			return err
		}
		log.Printf("Created the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.ResourceGroupName, obj.Region)
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
	if obj.HACluster {
		// HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
	} else {
		obj.AzureTokenCred = cred
		obj.Config = &AzureManagedState{}
		err := managedDeleteClusterHandler(ctx, obj)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(util.GetPath(util.CLUSTER_PATH, "azure", "managed", obj.ClusterName+" "+obj.ResourceGroupName+" "+obj.Region)); err != nil {
			return err
		}
		log.Printf("Deleted the cluster %s in resource group %s and region %s\n", obj.ClusterName, obj.ResourceGroupName, obj.Region)
	}
	return nil
}

func isPresent(kind string, obj AzureProvider) bool {
	path := util.GetPath(util.CLUSTER_PATH, "azure", kind, obj.ClusterName+" "+obj.ResourceGroupName+" "+obj.Region, "info.json")
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
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.ResourceGroupName}
			printKubeconfig.Printer(true, 0)
			return nil
		}
	case false:
		if isPresent("managed", provider) {
			var printKubeconfig util.PrinterKubeconfigPATH
			printKubeconfig = printer{ClusterName: provider.ClusterName, Region: provider.Region, ResourceName: provider.ResourceGroupName}
			printKubeconfig.Printer(false, 0)
			return nil
		}
	}
	return fmt.Errorf("ERR Cluster not found")
}
