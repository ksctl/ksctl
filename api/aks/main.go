/*
Kubesimplify
@maintainer:
*/

package aks

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kubesimplify/ksctl/api/utils"
	util "github.com/kubesimplify/ksctl/api/utils"
)


// fetchAPIKey returns the API key from the cred/civo file store
func fetchAPIKey() string {

	_, err := os.ReadFile(util.GetPath(0, "aks"))
	if err != nil {
		return ""
	}
	return ""
}

func Credentials() bool {
	// _, err := os.OpenFile(util.GetPath(0, "azure"), os.O_WRONLY, 0640)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	// sub ID, tenant id, client id, client secret,
	// export AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET
	skey := ""
	// tid := ""
	// pi := ""
	// pk := ""
	func() {
		fmt.Println("Enter your SUBSCRIPTION ID: ")
		_, err := fmt.Scan(&skey)
		if err != nil {
			panic(err.Error())
		}

		config := utils.AzureCredential{
			SubscriptionID: skey,
		}

		util.SaveCred(config, "azure")
	// 	fmt.Println("Enter your TENANT ID: ")
	// 	_, err = fmt.Scan(&tid)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	fmt.Println("Enter your SERVICE PRINCIPAL ID: ")
	// 	_, err = fmt.Scan(&pi)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	fmt.Println("Enter your : SERVICE PRINCIPAL KEY")
	// 	_, err = fmt.Scan(&pk)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	}()

	return true
	// _, err = file.Write([]byte(fmt.Sprintf(`Subscription-ID: %s
	// 	Tenant-ID: %s
	// 	Service-Principal-ID: %s
	// 	Service Principal-Key: %s`, skey, tid, pi, pk)))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	// return true
}


func (azure *AzureProvider) Create(cluster *Cluster, ha *HACluster, params map[string]string) error {
	// can we use same context for both the services i.e HA and managed cluster
	ctx := context.Background()

	// to validate the spec data and the credentials 
	if err := validate(azure); err != nil {
		return fmt.Errorf("caught up validating the Azure struct",err)
	}
	cred, err := getCredentials()
	if err != nil {
		return fmt.Errorf("caught getting the credentials:", err)
	}

	// TODO: implementation of details coming from cli
	// creating resource group
	subscriptionId, err := getSubscriptionId()
	if err != nil {
		fmt.Errorf("Caught up getting credentials", err)
	}

	if err := azure.getResource(ctx, params["region"], params["resourceGroupName"],cred, subscriptionId); err != nil {
		return fmt.Errorf("caught up creating the resource group", err)
	}
	// pass resource group name which on which we want to create the cluster 
	if cluster != nil {
		azure.Cluster = cluster
		if err := cluster.CreateCluster(ctx, cred, subscriptionId, params); err!=nil {
			return fmt.Errorf("caught up creating the cluster", err)
		}
	}

	return nil 
}

// create a resource group 
func (azure *AzureProvider) getResource(ctx context.Context, location string,resource string,cred *azidentity.DefaultAzureCredential, subscriptionID string) error {

	resourceClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	// setting up parameters for our resource group 
	param := armresources.ResourceGroup{
        Location: to.Ptr(location),
    }

    resourceGroup, err := resourceClient.CreateOrUpdate(ctx, resource, param, nil)
    if err != nil {
        return err
    }
	fmt.Println("Resource group created:", resource)
	// storing this with consideration that we might need this later 
	azure.ResourceGroups = resourceGroup
	return nil
}


func getSubscriptionId() (string, error) {
	sub, err := utils.GetCred("azure")
	if err != nil {
		return "", err
	}
	return sub["subscription_id"], nil
}

func getCredentials() (*azidentity.DefaultAzureCredential, error){
	// get credentials from the env varibles 
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Got Credentials")
	return cred, nil 
}

func validate(azure *AzureProvider) error {
	// Idea: search for os.GetEnv() and verify if they are present 
	// if they are not present the cred creation method also check for the cli cred(if cli is present)
	return nil 
}