package k3s

type Instances struct {
	ControlPlanes []string
	WorkerPlanes  []string
	DataStores    []string
	Loadbalancer  string
}

type StateConfiguration struct {
	K3sToken   string
	SSHUser    string
	PublicIPs  Instances
	PrivateIPs Instances
}

type K3sDistro struct {
	IsHA    bool
	Version string
}

// TODO: Add the SSH functionality here

var (
	k8sState *StateConfiguration
)
