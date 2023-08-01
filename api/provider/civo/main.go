package civo

import (
	"errors"
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

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
	ClusterName string     `json:"clustername"`
	Region      string     `json:"region"`
	SSHID       string     `json:"ssh_id"`
	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
	IPv4        InstanceIP `json:"ipv4_addr"`
}

var (
	currCloudState *StateConfiguration
)

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool
}

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	Application string `json:"application"`
	CNIPlugin   string `json:"cni_plugin"`
	Metadata
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (client *CivoProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (cloud.CloudResourceState, error) {
	payload := cloud.CloudResourceState{
		SSHState:          cloud.SSHPayload{PathPrivateKey: "abcd/rdcewcf"},
		Metadata:          cloud.Metadata{ClusterName: client.ClusterName},
		IPv4ControlPlanes: currCloudState.InstanceIDs.ControlNodes,
	}
	return payload, nil
}

// InitState implements resources.CloudInfrastructure.
func (*CivoProvider) InitState() error {
	if currCloudState != nil {
		return errors.New("[FATAL] already initialized")
	}
	currCloudState = &StateConfiguration{}
	currCloudState.InstanceIDs.ControlNodes = append(currCloudState.InstanceIDs.ControlNodes, "0.0.0.0")
	fmt.Println("[CIVO] Civo cloud state", currCloudState)
	return nil
}

func ReturnCivoStruct() *CivoProvider {
	return &CivoProvider{}
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudInfrastructure {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole string) resources.CloudInfrastructure {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudInfrastructure {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *CivoProvider) Visibility(toBePublic bool) resources.CloudInfrastructure {
	cloud.Metadata.Public = toBePublic
	return cloud
}
