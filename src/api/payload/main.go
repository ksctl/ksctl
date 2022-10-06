package payload

type Machine struct {
	Nodes uint8
	Cpu   uint8
	Mem   uint8
	Disk  uint8
}

type AwsProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
	AccessKey   string
	Secret      string
}

type AzureProvider struct {
	ClusterName         string
	HACluster           bool
	Region              string
	Spec                Machine
	SubscriptionID      string
	TenantID            string
	servicePrincipleKey string
	servicePrincipleID  string
}

type CivoProvider struct {
	ClusterName string
	APIKey      string
	HACluster   bool
	Region      string
	Spec        Machine
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
}

type Providers struct {
	eks  *AwsProvider
	aks  *AzureProvider
	k3s  *CivoProvider
	mk8s *LocalProvider
}
