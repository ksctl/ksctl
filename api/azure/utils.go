package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	util "github.com/kubesimplify/ksctl/api/utils"
)

type AzureProvider struct {
	ClusterName         string
	HACluster           bool
	Region              string
	Spec                util.Machine
	SubscriptionID      string
	TenantID            string
	ServicePrincipleKey string
	ServicePrincipleID  string
	ResourceGroups      armresources.ResourceGroupsClientCreateOrUpdateResponse //can remove it, temp
}

// type HACluster struct {
// 	HAClusterName string
// 	Region        string
// }

// type Cluster struct {
// 	ClusterName string
// 	Region      string
// 	NodeCount   int32
// }

// type Cred struct {
// 	SubscriptionID      string
// 	TenantID            string
// 	servicePrincipleKey string
// 	servicePrincipleID  string
// }
