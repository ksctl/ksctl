package aks

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kubesimplify/ksctl/api/utils"
)

type AzureProvider struct {
	User           string
	Cluster        *Cluster
	HACluster      *HACluster
	Spec           utils.Machine // do we really need this over here or keeping it inside cluster/HAcluster level makes more sense
	Credentials    *Cred
	ResourceGroups armresources.ResourceGroupsClientCreateOrUpdateResponse //can remove it, temp
}

type HACluster struct {
	HAClusterName string
	Region        string
}

type Cluster struct {
	ClusterName string
	Region      string
	NodeCount   int32
}

type Cred struct {
	SubscriptionID      string
	TenantID            string
	servicePrincipleKey string
	servicePrincipleID  string
}
