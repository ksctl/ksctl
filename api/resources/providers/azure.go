package providers

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type AzureProvider struct {
	ClusterName string `json:"cluster_name"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	// Spec           util.Machine `json:"spec"`
	SubscriptionID string `json:"subscription_id"`
	//Config         *AzureStateCluster     `json:"config"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"`
	//SSH_Payload    *util.SSHPayload       `json:"ssh___payload"`
}
