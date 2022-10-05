package payload

type Credential struct {
	AccessKey string
	Secret    string
}

type Machine struct {
	Nodes uint8
	Cpu   uint8
	Mem   uint8
}

type AwsProvider struct {
	ClusterName string
	Credential  Credential
	HACluster   bool
	Spec        Machine
}

type AzureProvider struct{}

type CivoProvider struct {
	ClusterName string
	Credential  Credential
	HACluster   bool
	Spec        Machine
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Spec        Machine
}

type Providers struct {
	eks  *AwsProvider
	aks  *AzureProvider
	k3s  *CivoProvider
	mk8s *LocalProvider
}
