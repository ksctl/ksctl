package azure

import (
	"errors"
	"fmt"

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

	// purpose: application in managed cluster
	Apps string
	Cni  string

	// these are used for managing the state and are the size of the arrays
	NoCP int
	NoWP int
	NoDS int
}

type AzureProvider struct {
	ClusterName   string `json:"cluster_name"`
	HACluster     bool   `json:"ha_cluster"`
	ResourceGroup string `json:"resource_group"`
	Region        string `json:"region"`
	// Spec           util.Machine `json:"spec"`
	SubscriptionID string `json:"subscription_id"`
	//Config         *AzureStateCluster     `json:"config"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"`
	//SSH_Payload    *util.SSHPayload       `json:"ssh___payload"`
	Metadata
}

// Version implements resources.CloudFactory.
func (*AzureProvider) Version(string) resources.CloudFactory {
	panic("unimplemented")
}

type Credential struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

var (
	azureCloudState *StateConfiguration
)

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (client *AzureProvider) CreateUploadSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelFirewall implements resources.CloudFactory.
func (*AzureProvider) DelFirewall(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelManagedCluster implements resources.CloudFactory.
func (*AzureProvider) DelManagedCluster(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelNetwork implements resources.CloudFactory.
func (*AzureProvider) DelNetwork(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudFactory.
func (*AzureProvider) DelSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelVM implements resources.CloudFactory.
func (*AzureProvider) DelVM(state resources.StorageFactory, indexNo int) error {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudFactory.
func (*AzureProvider) GetManagedKubernetes(state resources.StorageFactory) {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudFactory.
func (*AzureProvider) GetStateForHACluster(state resources.StorageFactory) (cloud.CloudResourceState, error) {
	panic("unimplemented")
}

// InitState implements resources.CloudFactory.
func (*AzureProvider) InitState(state resources.StorageFactory, operation string) error {
	if azureCloudState != nil {
		return errors.New("[FATAL] already initialized")
	}
	azureCloudState = &StateConfiguration{}
	return nil
}

// NewFirewall implements resources.CloudFactory.
func (*AzureProvider) NewFirewall(state resources.StorageFactory) error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudFactory.
func (*AzureProvider) NewManagedCluster(state resources.StorageFactory, noOfNodes int) error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudFactory.
func (*AzureProvider) NewNetwork(state resources.StorageFactory) error {
	panic("unimplemented")
}

// NewVM implements resources.CloudFactory.
func (*AzureProvider) NewVM(state resources.StorageFactory, indexNo int) error {
	return errors.New("unimplemented")
}

func ReturnAzureStruct(metadata resources.Metadata) *AzureProvider {
	return &AzureProvider{
		ClusterName:   metadata.ClusterName,
		Region:        metadata.Region,
		ResourceGroup: "", // TODO: add a field for resourse group need to be created, or check the main branch what created it
	}
}

// it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) resources.CloudFactory {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole string) resources.CloudFactory {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) resources.CloudFactory {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.Metadata.Public = toBePublic
	return cloud
}

// if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *AzureProvider) SupportForApplications() bool {
	return false
}

func (cloud *AzureProvider) SupportForCNI() bool {
	return false
}

func (client *AzureProvider) Application(s string) resources.CloudFactory {
	client.Metadata.Apps = s
	return client
}

func (client *AzureProvider) CNI(s string) resources.CloudFactory {
	client.Metadata.Cni = s
	return client
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *AzureProvider) NoOfControlPlane(no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.Metadata.NoCP = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *AzureProvider) NoOfDataStore(no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.Metadata.NoDS = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of Datastore>= 1 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (obj *AzureProvider) NoOfWorkerPlane(no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 0 {
		obj.Metadata.NoWP = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of workplane >= 0")
}
