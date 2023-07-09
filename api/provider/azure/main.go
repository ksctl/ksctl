package azure

import "github.com/Azure/azure-sdk-for-go/sdk/azcore"

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

func (a *AzureProvider) CreateVM() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) DeleteVM() {
	//TODO implement me
	panic("implement me")
}

func (a AzureProvider) CreateFirewall() {
	//TODO implement me
	panic("implement me")
}

func (a AzureProvider) DeleteFirewall() {
	//TODO implement me
	panic("implement me")
}

func (a AzureProvider) CreateVirtualNetwork() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) DeleteVirtualNetwork() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) GetVM() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) CreateManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) GetManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}

func (a *AzureProvider) DeleteManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}
