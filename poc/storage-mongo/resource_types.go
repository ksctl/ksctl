package main

import (
	"github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageDocument struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	// composite primary key
	ClusterType string `json:"cluster_type" bson:"cluster_type" `
	Region      string `json:"region" bson:"region"`
	ClusterName string `json:"cluster_name" bson:"cluster_name"`

	CloudInfra      *InfrastructureState      `json:"cloud_infrastructure_state" bson:"cloud_infrastructure_state,omitempty"`
	BootStrapConfig *KubernetesBootstrapState `json:"kubernetes_bootstrap_state" bson:"kubernetes_bootstrap_state,omitempty"`

	ClusterKubeConfig string `json:"cluster_kubeconfig" bson:"cluster_kubeconfig"` // TODO: should we use k8s.io/kops/pkg/kubeconfig or only string

	// ClusterKubeConfig kubeconfig.KubectlConfig `json:"cluster_kubeconfig" bson:"cluster_kubeconfig"`

	SSHKeyPair SSHKeyPairState `json:"ssh_key_pair" bson:"ssh_key_pair"`
}

type InfrastructureState struct {
	Azure *AzureState `json:"azure" bson:"azure,omitempty"`
	Civo  *CivoState  `json:"civo" bson:"civo,omitempty"`
}

type KubernetesBootstrapState struct {
	K3s *K3sBootstrapState `json:"k3s" bson:"k3s,omitempty"`
}

type SSHKeyPairState struct {
	PublicKey  string `json:"public_key" bson:"public_key"`
	PrivateKey string `json:"private_key" bson:"private_key"`
}

type K3sBootstrapState struct {
	K3sToken          string        `json:"k3s_token" bson:"k3s_token"`
	DataStoreEndPoint string        `json:"datastore_endpoint" bson:"datastore_endpoint"`
	SSHInfo           cloud.SSHInfo `json:"cloud_ssh_info" bson:"cloud_ssh_info"`
	PublicIPs         Instances     `json:"cloud_public_ips" bson:"cloud_public_ips"`
	PrivateIPs        Instances     `json:"cloud_private_ips" bson:"cloud_private_ips"`

	ClusterName string                  `json:"cluster_name" bson:"cluster_name"`
	Region      string                  `json:"region" bson:"region"`
	ClusterType consts.KsctlClusterType `json:"cluster_type" bson:"cluster_type"`
	ClusterDir  string                  `json:"cluster_dir" bson:"cluster_dir"`
	Provider    consts.KsctlCloud       `json:"provider" bson:"provider"`
}

// specific to each infrastructure like civo or azure
type AzureState struct {
	IsCompleted bool `json:"status" bson:"status"`

	ResourceGroupName string `json:"resource_group_name" bson:"resource_group_name"`

	SSHUser          string `json:"ssh_usr" bson:"ssh_usr"`
	SSHPrivateKeyLoc string `json:"ssh_private_key_location" bson:"ssh_private_key_location"` // TODO: we need to remove the dependency of paths
	SSHKeyName       string `json:"sshkey_name" bson:"sshkey_name"`

	// ManagedCluster
	ManagedClusterName string `json:"managed_cluster_name" bson:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_cluster_nodes" bson:"no_managed_cluster_nodes"`

	SubnetName         string        `json:"subnet_name" bson:"subnet_name"`
	SubnetID           string        `json:"subnet_id" bson:"subnet_id"`
	VirtualNetworkName string        `json:"virtual_network_name" bson:"virtual_network_name"`
	VirtualNetworkID   string        `json:"virtual_network_id" bson:"virtual_network_id"`
	InfoControlPlanes  AzureStateVMs `json:"info_control_planes" bson:"info_control_planes"`
	InfoWorkerPlanes   AzureStateVMs `json:"info_worker_planes" bson:"info_worker_planes"`
	InfoDatabase       AzureStateVMs `json:"info_database" bson:"info_database"`
	InfoLoadBalancer   AzureStateVM  `json:"info_load_balancer" bson:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro" bson:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version" bson:"k8s_version"`
}

type CivoState struct {
	IsCompleted bool `json:"status" bson:"status"`

	NetworkName string `json:"resource_group_name" bson:"resource_group_name"`
}

// specific to each infrastructure like civo or azure
type Instances struct {
	ControlPlanes []string `json:"controlplanes" bson:"controlplanes"`
	WorkerPlanes  []string `json:"workerplanes" bson:"workerplanes"`
	DataStores    []string `json:"datastores" bson:"datastores"`
	Loadbalancer  string   `json:"loadbalancer" bson:"loadbalancer"`
}

// specific to each infrastructure like civo or azure
type AzureStateVMs struct {
	Names                    []string `json:"names" bson:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name" bson:"network_security_group_name"`
	NetworkSecurityGroupID   string   `json:"network_security_group_id" bson:"network_security_group_id"`
	DiskNames                []string `json:"disk_names" bson:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names" bson:"public_ip_names"`
	PublicIPIDs              []string `json:"public_ip_ids" bson:"public_ip_ids"`
	PrivateIPs               []string `json:"private_ips" bson:"private_ips"`
	PublicIPs                []string `json:"public_ips" bson:"public_ips"`
	NetworkInterfaceNames    []string `json:"network_interface_names" bson:"network_interface_names"`
	NetworkInterfaceIDs      []string `json:"network_interface_ids" bson:"network_interface_ids"`
	Hostnames                []string `json:"hostnames" bson:"hostnames"`
}

type AzureStateVM struct {
	Name                     string `json:"name" bson:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name" bson:"network_security_group_name"`
	NetworkSecurityGroupID   string `json:"network_security_group_id" bson:"network_security_group_id"`
	DiskName                 string `json:"disk_name" bson:"disk_name"`
	PublicIPName             string `json:"public_ip_name" bson:"public_ip_name"`
	PublicIPID               string `json:"public_ip_id" bson:"public_ip_id"`
	NetworkInterfaceName     string `json:"network_interface_name" bson:"network_interface_name"`
	NetworkInterfaceID       string `json:"network_interface_id" bson:"network_interface_id"`
	PrivateIP                string `json:"private_ip" bson:"private_ip"`
	PublicIP                 string `json:"public_ip" bson:"public_ip"`
	HostName                 string `json:"hostname" bson:"hostname"`
}
