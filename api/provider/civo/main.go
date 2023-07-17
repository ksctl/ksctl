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

type StateConfiguration struct {
	ClusterName string     `json:"clustername"`
	Region      string     `json:"region"`
	DBEndpoint  string     `json:"dbendpoint"`
	ServerToken string     `json:"servertoken"`
	SSHID       string     `json:"ssh_id"`
	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
}

type CloudController cloud.ClientBuilder

func WrapCloudControllerBuilder(b *cloud.ClientBuilder) *CloudController {
	civo := (*CloudController)(b)
	return civo
}

func (client *CloudController) CreateHACluster() {

	fmt.Println("Implement me[civo ha create]")
	err := client.State.Save("civo.txt", nil)
	fmt.Println(err)
	client.Distro.ConfigureControlPlane()
}

func (client *CloudController) CreateManagedCluster() {
	fmt.Println("Implement me[civo managed create]")

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
