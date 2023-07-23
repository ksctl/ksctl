package civo

import (
	"fmt"

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
	ClusterName string                   `json:"clustername"`
	Region      string                   `json:"region"`
	SSHID       string                   `json:"ssh_id"`
	InstanceIDs InstanceID               `json:"instanceids"`
	NetworkIDs  NetworkID                `json:"networkids"`
	IPv4        InstanceIP               `json:"ipv4_addr"`
	K8s         cloud.CloudResourceState // dont include it here it should be present in kubernetes
}

type CloudController cloud.ClientBuilder

var (
	currCloudState *StateConfiguration
)

// FetchState implements cloud.ControllerInterface.
func (*CloudController) FetchState() cloud.CloudResourceState {
	return currCloudState.K8s
}

func WrapCloudControllerBuilder(b *cloud.ClientBuilder) *CloudController {
	civo := (*CloudController)(b)
	return civo
}

func (client *CloudController) CreateHACluster() {

	fmt.Println("Implement me[civo ha create]")
	currCloudState = &StateConfiguration{
		ClusterName: client.ClusterName,
		Region:      client.Region,
		K8s: cloud.CloudResourceState{
			SSHState: cloud.SSHPayload{UserName: "root"},
			Metadata: cloud.Metadata{
				ClusterName: client.ClusterName,
				Region:      client.Region,
				Provider:    "civo",
			},
		},
	}
	err := client.State.Save("civo.txt", nil)
	fmt.Println(err)
	client.Distro.ConfigureControlPlane()
}

func (client *CloudController) CreateManagedCluster() {
	fmt.Println("Implement me[civo managed create]")

	currCloudState = nil
	currCloudState = &StateConfiguration{
		ClusterName: client.ClusterName,
	}
	client.Cloud.CreateManagedKubernetes()

	_, err := client.State.Load("civo.txt")
	fmt.Println(err)
}

func (client *CloudController) DestroyHACluster() {
	fmt.Println("Implement me[civo ha delete]")
}

func (client *CloudController) DestroyManagedCluster() {

	fmt.Println("Implement me[civo managed delete]")
}
