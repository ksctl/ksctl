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
	ClusterName    string
	HACluster      bool
	Region         string
	Spec           util.Machine
	SubscriptionID string
	// TenantID            string
	// ServicePrincipleKey string
	// ServicePrincipleID  string
	ResourceGroupName string
	AzureTokenCred    azcore.TokenCredential
}

func (obj *AzureProvider) CreateCluster() {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if obj.HACluster {
		// HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
	} else {
		obj.AzureTokenCred = cred
		managedCluster, err := managedCreateClusterHandler(ctx, obj)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("managed cluster created: ", *managedCluster.ID)

	}

}

func (obj *AzureProvider) DeleteCluster() {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if obj.HACluster {
		// HA CLUSTER CREATE
		log.Println("TO BE DEVELOPED")
	} else {
		obj.AzureTokenCred = cred
		managedCluster, err := managedDeleteClusterHandler(ctx, obj)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("managed cluster deleted: ", *managedCluster.ID)
	}
}
