package civo

import (
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	//Spec        util.Machine `json:"spec"`
	Application string `json:"application"`
	CNIPlugin   string `json:"cni_plugin"`
}

type InstanceID struct {
	ControlNodes     []string `json:"controlnodeids"`
	WorkerNodes      []string `json:"workernodeids"`
	LoadBalancerNode []string `json:"loadbalancernodeids"`
	DatabaseNode     []string `json:"databasenodeids"`
}

type NetworkID struct {
	FirewallIDControlPlaneNode string `json:"fwidcontrolplanenode"`
	FirewallIDWorkerNode       string `json:"fwidworkernode"`
	FirewallIDLoadBalancerNode string `json:"fwidloadbalancenode"`
	FirewallIDDatabaseNode     string `json:"fwiddatabasenode"`
	NetworkID                  string `json:"clusternetworkid"`
}

type InstanceIP struct {
	IPControlplane        []string
	IPWorkerPlane         []string
	IPLoadbalancer        string
	IPDataStore           []string
	PrivateIPControlplane []string
	PrivateIPWorkerPlane  []string
	PrivateIPLoadbalancer string
	PrivateIPDataStore    []string
}

type StateConfiguration struct {
	ClusterName string                   `json:"clustername"`
	Region      string                   `json:"region"`
	SSHID       string                   `json:"ssh_id"`
	InstanceIDs InstanceID               `json:"instanceids"`
	NetworkIDs  NetworkID                `json:"networkids"`
	IPv4        InstanceIP               `json:"ipv4_addr"`
	K8s         cloud.CloudResourceState // dont include it here it should be present in kubernetes
	// for HA different StateConfiguration and for Managed different StateConfiguration
	// HA cloud.CloudResourceState
	// Managed Cmanaed
}

// type CloudController cloud.ClientBuilder

// var (
// 	currCloudState *StateConfiguration
// )

// // FetchState implements cloud.ControllerInterface.
// func (*CloudController) FetchState() cloud.CloudResourceState {
// 	return currCloudState.K8s
// }

// func WrapCloudControllerBuilder(b *cloud.ClientBuilder) *CloudController {
// 	civo := (*CloudController)(b)
// 	return civo
// }
