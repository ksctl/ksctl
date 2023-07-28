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

// Plan is to have seperate fileds inside it so that it will be much easier to transfer data
// ex. ha state and managed state
// when managed state is going to be used the specific section need to be used
// other one should be null
type StateConfiguration struct {
	ClusterName string                   `json:"clustername"`
	Region      string                   `json:"region"`
	SSHID       string                   `json:"ssh_id"`
	InstanceIDs InstanceID               `json:"instanceids"`
	NetworkIDs  NetworkID                `json:"networkids"`
	IPv4        InstanceIP               `json:"ipv4_addr"`
	K8s         cloud.CloudResourceState // dont include it here it should be present in kubernetes
	// for HA different StateConfiguration and for Managed different StateConfiguration
	// WARN: HA cloud.CloudResourceState
	// WARN: Managed xyz..
}

// type CloudController cloud.ClientBuilder

var (
	currCloudState *StateConfiguration
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

// CreateUploadSSHKeyPair implements resources.CloudInfrastructure.
func (*CivoProvider) CreateUploadSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelFirewall implements resources.CloudInfrastructure.
func (*CivoProvider) DelFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (*CivoProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelNetwork implements resources.CloudInfrastructure.
func (*CivoProvider) DelNetwork(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudInfrastructure.
func (*CivoProvider) DelSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelVM implements resources.CloudInfrastructure.
func (*CivoProvider) DelVM(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (*CivoProvider) GetManagedKubernetes(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (client *CivoProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (cloud.CloudResourceState, error) {
	payload := cloud.CloudResourceState{
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
	fmt.Println("Civo cloud state", currCloudState)
	return nil
}

// NewFirewall implements resources.CloudInfrastructure.
func (*CivoProvider) NewFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudInfrastructure.
func (*CivoProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudInfrastructure.
func (*CivoProvider) NewNetwork(state resources.StateManagementInfrastructure) error {
	fmt.Println("[CIVO] Creating network...")
	return state.Save("civoNet.txt", nil)
	// return nil
}

// NewVM implements resources.CloudInfrastructure.
func (*CivoProvider) NewVM(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

func ReturnCivoStruct() *CivoProvider {
	return &CivoProvider{}
}
