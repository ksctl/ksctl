package azure

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kubesimplify/ksctl/api/resources"
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
	ClusterName        string                   `json:"cluster_name"`
	Region             string                   `json:"region"`
	ResourceGroupName  string                   `json:"resource_group_name"`
	SSHKeyName         string                   `json:"ssh_key_name"`
	SubnetName         string                   `json:"subnet_name"`
	SubnetID           string                   `json:"subnet_id"`
	VirtualNetworkName string                   `json:"virtual_network_name"`
	VirtualNetworkID   string                   `json:"virtual_network_id"`
	InfoControlPlanes  AzureStateVMs            `json:"info_control_planes"`
	InfoWorkerPlanes   AzureStateVMs            `json:"info_worker_planes"`
	InfoDatabase       AzureStateVM             `json:"info_database"`
	InfoLoadBalancer   AzureStateVM             `json:"info_load_balancer"`
	K8s                cloud.CloudResourceState // dont include it here it should be present in kubernetes
}
type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool
}

type AzureProvider struct {
	ClusterName string `json:"cluster_name"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	// Spec           util.Machine `json:"spec"`
	SubscriptionID string `json:"subscription_id"`
	//Config         *AzureStateCluster     `json:"config"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"`
	//SSH_Payload    *util.SSHPayload       `json:"ssh___payload"`
	Metadata
}

var (
	currCloudState *StateConfiguration
)

// CreateUploadSSHKeyPair implements resources.CloudInfrastructure.
func (client *AzureProvider) CreateUploadSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelFirewall implements resources.CloudInfrastructure.
func (*AzureProvider) DelFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (*AzureProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelNetwork implements resources.CloudInfrastructure.
func (*AzureProvider) DelNetwork(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudInfrastructure.
func (*AzureProvider) DelSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelVM implements resources.CloudInfrastructure.
func (*AzureProvider) DelVM(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (*AzureProvider) GetManagedKubernetes(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (*AzureProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (cloud.CloudResourceState, error) {
	panic("unimplemented")
}

// InitState implements resources.CloudInfrastructure.
func (*AzureProvider) InitState() error {
	if currCloudState != nil {
		return errors.New("[FATAL] already initialized")
	}
	currCloudState = &StateConfiguration{}
	return nil
}

// NewFirewall implements resources.CloudInfrastructure.
func (*AzureProvider) NewFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudInfrastructure.
func (*AzureProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudInfrastructure.
func (*AzureProvider) NewNetwork(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewVM implements resources.CloudInfrastructure.
func (*AzureProvider) NewVM(state resources.StateManagementInfrastructure) error {
	_ = state.Save("abcd.txt", nil)
	return errors.New("unimplemented")
}

func ReturnAzureStruct() *AzureProvider {
	return &AzureProvider{}
}

// it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) resources.CloudInfrastructure {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole string) resources.CloudInfrastructure {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) resources.CloudInfrastructure {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) resources.CloudInfrastructure {
	cloud.Metadata.Public = toBePublic
	return cloud
}
