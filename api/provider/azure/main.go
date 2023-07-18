package azure

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

// IMPORTANT: the state management structs are local to each provider thus making each of them unique
// but the problem is we need to pass some required values from the cloud providers to the kubernetesdistro
// but how?
// can we use the controllers as a bridge to allow it to happen when we are going to transfer the resources
// if this is the case we need to figure out the way to do so
// also figure out, where the stateConfiguration struct vairable be present (i.e. in controller or inside this?)

type AzureStateVMs struct {
	Names                    []string `json:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name"`
	NetworkSecurityGroupID   string   `json:"network_security_group_id"`
	DiskNames                []string `json:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names"`
	PrivateIPs               []string `json:"private_ips"`
	PublicIPs                []string `json:"public_ips"`
	NetworkInterfaceNames    []string `json:"network_interface_names"`
}

type AzureStateVM struct {
	Name                     string `json:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name"`
	NetworkSecurityGroupID   string `json:"network_security_group_id"`
	DiskName                 string `json:"disk_name"`
	PublicIPName             string `json:"public_ip_name"`
	NetworkInterfaceName     string `json:"network_interface_name"`
	PrivateIP                string `json:"private_ip"`
	PublicIP                 string `json:"public_ip"`
}

type StateConfiguration struct {
	ClusterName       string `json:"cluster_name"`
	ResourceGroupName string `json:"resource_group_name"`
	SSHKeyName        string `json:"ssh_key_name"`
	// DBEndpoint        string `json:"database_endpoint"`
	// K3sToken          string `json:"k3s_token"`

	SubnetName         string `json:"subnet_name"`
	SubnetID           string `json:"subnet_id"`
	VirtualNetworkName string `json:"virtual_network_name"`
	VirtualNetworkID   string `json:"virtual_network_id"`

	InfoControlPlanes AzureStateVMs `json:"info_control_planes"`
	InfoWorkerPlanes  AzureStateVMs `json:"info_worker_planes"`
	InfoDatabase      AzureStateVM  `json:"info_database"`
	InfoLoadBalancer  AzureStateVM  `json:"info_load_balancer"`
}

type CloudController cloud.ClientBuilder

func WrapCloudControllerBuilder(b *cloud.ClientBuilder) *CloudController {
	azure := (*CloudController)(b)
	return azure
}

func (client *CloudController) CreateHACluster() {

	fmt.Println("Implement me[azure ha create]")
	err := client.State.Save("azure.txt", nil)
	fmt.Println(err)
	client.Distro.ConfigureControlPlane()
}

func (client *CloudController) CreateManagedCluster() {
	fmt.Println("Implement me[azure managed create]")

	client.Cloud.CreateManagedKubernetes()

	_, err := client.State.Load("azure.txt")
	fmt.Println(err)
}

func (client *CloudController) DestroyHACluster() {
	fmt.Println("Implement me[azure ha delete]")
}

func (client *CloudController) DestroyManagedCluster() {

	fmt.Println("Implement me[azure managed delete]")
}
