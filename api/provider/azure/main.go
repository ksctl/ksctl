package azure

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

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
type HostNames struct {
	ControlNodes     []string `json:"controlnode"`
	WorkerNodes      []string `json:"workernode"`
	LoadBalancerNode string   `json:"loadbalancernode"`
	DatabaseNode     []string `json:"databasenodes"`
}

type StateConfiguration struct {
	IsCompleted bool `json:"status"`

	ClusterName       string `json:"cluster_name"`
	Region            string `json:"region"`
	ResourceGroupName string `json:"resource_group_name"`

	SSHID            string `json:"ssh_id"`
	SSHUser          string `json:"ssh_usr"`
	SSHPrivateKeyLoc string `json:"ssh_private_key_location"`

	SubnetName         string        `json:"subnet_name"`
	SubnetID           string        `json:"subnet_id"`
	VirtualNetworkName string        `json:"virtual_network_name"`
	VirtualNetworkID   string        `json:"virtual_network_id"`
	InfoControlPlanes  AzureStateVMs `json:"info_control_planes"`
	InfoWorkerPlanes   AzureStateVMs `json:"info_worker_planes"`
	InfoDatabase       AzureStateVMs `json:"info_database"`
	InfoLoadBalancer   AzureStateVM  `json:"info_load_balancer"`
}

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool

	// purpose: application in managed cluster
	Apps    string
	Cni     string
	Version string

	// these are used for managing the state and are the size of the arrays
	NoCP int
	NoWP int
	NoDS int

	K8sName    string
	K8sVersion string
}

type AzureProvider struct {
	ClusterName    string                 `json:"cluster_name"`
	HACluster      bool                   `json:"ha_cluster"`
	ResourceGroup  string                 `json:"resource_group"`
	Region         string                 `json:"region"`
	SubscriptionID string                 `json:"subscription_id"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"`
	SSHPath        string                 `json:"ssh_key"`
	Metadata
}

var (
	azCloudState *StateConfiguration
	//azClient     *civogo.Client
	clusterDirName string
	clusterType    string // it stores the ha or managed

)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

func (*AzureProvider) GetHostNameAllWorkerNode() []string {
	//TODO implement me
	panic("implement me")
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

func ReturnAzureStruct(metadata resources.Metadata) *AzureProvider {
	return &AzureProvider{
		ClusterName:   metadata.ClusterName,
		Region:        metadata.Region,
		ResourceGroup: "", // TODO: add a field for resourse group need to be created, or check the main branch what created it
	}
}

// Name it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) resources.CloudFactory {
	cloud.Metadata.ResName = resName
	return cloud
}

// Role it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole string) resources.CloudFactory {
	cloud.Metadata.Role = resRole
	return cloud
}

// VMType it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) resources.CloudFactory {
	cloud.Metadata.VmType = size
	return cloud
}

// Visibility whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.Metadata.Public = toBePublic
	return cloud
}

// SupportForApplications if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *AzureProvider) SupportForApplications() bool {
	return false
}

func (cloud *AzureProvider) SupportForCNI() bool {
	return false
}

func (cloud *AzureProvider) Application(s string) resources.CloudFactory {
	cloud.Metadata.Apps = s
	return cloud
}

func (cloud *AzureProvider) CNI(s string) resources.CloudFactory {
	cloud.Metadata.Cni = s
	return cloud
}

// NoOfControlPlane implements resources.CloudFactory.
func (cloud *AzureProvider) NoOfControlPlane(no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 3 && (no&1) == 1 {
		cloud.Metadata.NoCP = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (cloud *AzureProvider) NoOfDataStore(no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 1 && (no&1) == 1 {
		cloud.Metadata.NoDS = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of Datastore>= 1 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (cloud *AzureProvider) NoOfWorkerPlane(factory resources.StorageFactory, no int, isCreateOperation bool) (int, error) {
	if !isCreateOperation {
		return 0, nil
	}
	if no >= 0 {
		cloud.Metadata.NoWP = no
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of workplane >= 0")
}
