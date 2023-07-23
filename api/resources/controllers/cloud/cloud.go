package cloud

import "github.com/kubesimplify/ksctl/api/resources"

const (
	CREATE = 0
	DELETE = 1
	GET    = 2
)

type ClientBuilder resources.Builder

type ClientStateMgt resources.StateManagementInfrastructure

// CloudResourceState provides the state which cloud provider creates
// and which is consumed by the kubernetes to configure them
type CloudResourceState struct {
	SSHState          SSHPayload
	IPv4ControlPlanes []string
	IPv4WorkerPlanes  []string
	IPv4DataStores    []string
	IPv4LoadBalancer  string

	PrivateIPv4ControlPlanes []string
	PrivateIPv4WorkerPlanes  []string
	PrivateIPv4DataStores    []string
	PrivateIPv4LoadBalancer  string
	Metadata                 Metadata
}

type Metadata struct {
	ClusterName   string
	Region        string  // for the civo
	ResourceGroup *string // for azure // CHECK: if its required
	VPC           *string // for aws // CHECK: if its required
	Provider      string
}

type SSHPayload struct {
	UserName       string
	PathPrivateKey string
	Output         string
}

type ControllerInterface interface {
	CreateHACluster()
	CreateManagedCluster()

	DestroyHACluster()
	DestroyManagedCluster()

	FetchState() CloudResourceState
}
