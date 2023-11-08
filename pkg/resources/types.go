package resources

import (
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"io"
)

type KsctlClient struct {
	// Cloud is the CloudProvider's factory interface
	Cloud CloudFactory

	// Distro is the Distrobution's factory interface
	Distro DistroFactory

	// Storage is the Storage's factory interface
	Storage StorageFactory

	// Metadata is used by the cloudController and manager to use data from cli
	Metadata Metadata
}

// this is used against the cloud provider meaning it is used to store the state of cloud provider
type Metadata struct {
	ClusterName string `json:"cluster_name"`
	Region      string `json:"region"`

	Provider      consts.KsctlCloud      `json:"cloud_provider"`
	K8sDistro     consts.KsctlKubernetes `json:"kubernetes_distro"`
	StateLocation consts.KsctlStore      `json:"storage_type"`

	IsHA bool `json:"ha_cluster"`

	K8sVersion string `json:"kubernetes_version"`

	ManagedNodeType      string `json:"node_type_managed"`
	WorkerPlaneNodeType  string `json:"node_type_workerplane"`
	ControlPlaneNodeType string `json:"node_type_controlplane"`
	DataStoreNodeType    string `json:"node_type_datastore"`
	LoadBalancerNodeType string `json:"node_type_loadbalancer"`

	NoMP int `json:"desired_no_of_managed_nodes"`      // No of managed Nodes
	NoWP int `json:"desired_no_of_workerplane_nodes"`  // No of woerkplane VMs
	NoCP int `json:"desired_no_of_controlplane_nodes"` // No of Controlplane VMs
	NoDS int `json:"desired_no_of_datastore_nodes"`    // No of DataStore VMs

	Applications string `json:"preinstalled_apps"`
	CNIPlugin    string `json:"cni_plugin"`

	LogVerbosity int       `json:"log_verbosity"`
	LogWritter   io.Writer `json:"log_writter"`
}

//// CobraCmd TODO: Move it to the cli repo
//type CobraCmd struct {
//	ClusterName string
//	Region      string
//	Client      KsctlClient
//	Version     string
//}
