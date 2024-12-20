package provisioner

import (
	"github.com/ksctl/ksctl/pkg/bootstrap"
	"github.com/ksctl/ksctl/pkg/bootstrap/distributions"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/providers"
	"github.com/ksctl/ksctl/pkg/storage"
)

type Client struct {
	Cloud providers.Cloud

	PreBootstrap bootstrap.Bootstrap

	Bootstrap distributions.KubernetesDistribution

	Storage storage.Storage

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

	//Applications []KsctlApp `json:"preinstalled_apps"`
	//CNIPlugin    KsctlApp   `json:"cni_plugin"`
}

//type KsctlApp struct {
//	StackName string                    `json:"stack_name"`
//	Overrides map[string]map[string]any `json:"overrides"`
//}
