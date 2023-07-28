package kubeadm

// OTHER CONFIGURATIONS
type Instances struct {
	ControlPlanes []string
	WorkerPlanes  []string
	DataStores    []string
	Loadbalancer  string
}

type StateConfiguration struct {
	JoinControlToken string
	JoinWorkerToken  string
	SSHUser          string
	PublicIPs        Instances
	PrivateIPs       Instances
}

type KubeadmDistro struct {
	IsHA    bool
	Version string
}

var (
	k8sState *StateConfiguration
)
