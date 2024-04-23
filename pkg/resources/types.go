package resources

import (
	"io"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type KsctlClient struct {
	Cloud CloudFactory

	PreBootstrap PreKubernetesBootstrap

	Bootstrap KubernetesBootstrap

	Storage StorageFactory

	Metadata Metadata
}

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

	Applications []string `json:"preinstalled_apps"`
	CNIPlugin    string   `json:"cni_plugin"`

	LogVerbosity int       `json:"log_verbosity"`
	LogWritter   io.Writer `json:"log_writter"`
}
