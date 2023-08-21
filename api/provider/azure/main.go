package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kubesimplify/ksctl/api/logger"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/kubesimplify/ksctl/api/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
)

type AzureStateVMs struct {
	Names                    []string `json:"names"`
	NetworkSecurityGroupName string   `json:"network_security_group_name"`
	NetworkSecurityGroupID   string   `json:"network_security_group_id"`
	DiskNames                []string `json:"disk_names"`
	PublicIPNames            []string `json:"public_ip_names"`
	PublicIPIDs              []string `json:"public_ip_ids"`
	PrivateIPs               []string `json:"private_ips"`
	PublicIPs                []string `json:"public_ips"`
	NetworkInterfaceNames    []string `json:"network_interface_names"`
	NetworkInterfaceIDs      []string `json:"network_interface_ids"`
	Hostnames                []string `json:"hostnames"`
}

type AzureStateVM struct {
	Name                     string `json:"name"`
	NetworkSecurityGroupName string `json:"network_security_group_name"`
	NetworkSecurityGroupID   string `json:"network_security_group_id"`
	DiskName                 string `json:"disk_name"`
	PublicIPName             string `json:"public_ip_name"`
	PublicIPID               string `json:"public_ip_id"`
	NetworkInterfaceName     string `json:"network_interface_name"`
	NetworkInterfaceID       string `json:"network_interface_id"`
	PrivateIP                string `json:"private_ip"`
	PublicIP                 string `json:"public_ip"`
	HostName                 string `json:"hostname"`
}

type StateConfiguration struct {
	IsCompleted bool `json:"status"`

	ClusterName       string `json:"cluster_name"`
	Region            string `json:"region"`
	ResourceGroupName string `json:"resource_group_name"`

	// SSHID            string `json:"ssh_id"`
	SSHUser          string `json:"ssh_usr"`
	SSHPrivateKeyLoc string `json:"ssh_private_key_location"`
	SSHKeyName       string `json:"sshkey_name"`

	// ManagedCluster
	ManagedClusterName string `json:"managed_cluster_name"`
	NoManagedNodes     int    `json:"no_managed_cluster_nodes"`

	SubnetName         string        `json:"subnet_name"`
	SubnetID           string        `json:"subnet_id"`
	VirtualNetworkName string        `json:"virtual_network_name"`
	VirtualNetworkID   string        `json:"virtual_network_id"`
	InfoControlPlanes  AzureStateVMs `json:"info_control_planes"`
	InfoWorkerPlanes   AzureStateVMs `json:"info_worker_planes"`
	InfoDatabase       AzureStateVMs `json:"info_database"`
	InfoLoadBalancer   AzureStateVM  `json:"info_load_balancer"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
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
	ClusterName   string `json:"cluster_name"`
	HACluster     bool   `json:"ha_cluster"`
	ResourceGroup string `json:"resource_group"`
	Region        string `json:"region"`

	// DEPRICATION: move to the az client
	SubscriptionID string                 `json:"subscription_id"`
	AzureTokenCred azcore.TokenCredential `json:"azure_token_cred"
`
	SSHPath  string `json:"ssh_key"`
	Metadata Metadata

	Client AzureGo
}

var (
	azureCloudState *StateConfiguration

	clusterDirName string
	clusterType    string // it stores the ha or managed

	ctx context.Context
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

func (*AzureProvider) GetHostNameAllWorkerNode() []string {
	var hostnames []string = make([]string, len(azureCloudState.InfoWorkerPlanes.Hostnames))
	copy(hostnames, azureCloudState.InfoWorkerPlanes.Hostnames)
	return hostnames
}

// Version implements resources.CloudFactory.
func (obj *AzureProvider) Version(ver string) resources.CloudFactory {
	if err := isValidK8sVersion(obj, ver); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}

	obj.Metadata.K8sVersion = ver
	return obj
}

type Credential struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

// GetManagedKubernetes implements resources.CloudFactory.
func (*AzureProvider) GetManagedKubernetes(resources.StorageFactory) {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudFactory.
// WARN: the array copy is a shallow copy
func (*AzureProvider) GetStateForHACluster(storage resources.StorageFactory) (cloud_control_res.CloudResourceState, error) {
	payload := cloud_control_res.CloudResourceState{
		SSHState: cloud_control_res.SSHInfo{
			PathPrivateKey: azureCloudState.SSHPrivateKeyLoc,
			UserName:       azureCloudState.SSHUser,
		},
		Metadata: cloud_control_res.Metadata{
			ClusterName: azureCloudState.ClusterName,
			Provider:    "azure",
			Region:      azureCloudState.Region,
			ClusterType: clusterType,
			ClusterDir:  clusterDirName,
		},
		// Public IPs
		IPv4ControlPlanes: azureCloudState.InfoControlPlanes.PublicIPs,
		IPv4DataStores:    azureCloudState.InfoDatabase.PublicIPs,
		IPv4WorkerPlanes:  azureCloudState.InfoWorkerPlanes.PublicIPs,
		IPv4LoadBalancer:  azureCloudState.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: azureCloudState.InfoControlPlanes.PrivateIPs,
		PrivateIPv4DataStores:    azureCloudState.InfoDatabase.PrivateIPs,
		PrivateIPv4LoadBalancer:  azureCloudState.InfoLoadBalancer.PrivateIP,
	}
	storage.Logger().Success("[azure] Transferred Data, it's ready to be shipped!")
	return payload, nil
}

// InitState implements resources.CloudFactory.
func (obj *AzureProvider) InitState(storage resources.StorageFactory, operation string) error {

	switch obj.HACluster {
	case false:
		clusterType = utils.CLUSTER_TYPE_MANG
	case true:
		clusterType = utils.CLUSTER_TYPE_HA
	}
	obj.ResourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.ClusterName, clusterType)
	clusterDirName = obj.ClusterName + " " + obj.ResourceGroup + " " + obj.Region

	if azureCloudState != nil {
		return errors.New("[FATAL] already initialized")
	}
	errLoadState := loadStateHelper(storage)
	switch operation {
	case utils.OPERATION_STATE_CREATE:
		if errLoadState == nil && azureCloudState.IsCompleted {
			return fmt.Errorf("[azure] already exist")
		}
		if errLoadState == nil && !azureCloudState.IsCompleted {
			storage.Logger().Note("[azure] RESUME triggered!!")
		} else {
			storage.Logger().Note("[azure] Fresh state!!")
			azureCloudState = &StateConfiguration{
				IsCompleted:      false,
				ClusterName:      obj.ClusterName,
				Region:           obj.Region,
				KubernetesDistro: obj.Metadata.K8sName,
				KubernetesVer:    obj.Metadata.K8sVersion,
			}
		}

	case utils.OPERATION_STATE_DELETE:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[azure] Delete resource(s)")

	case utils.OPERATION_STATE_GET:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[azure] Get resources")
		clusterDirName = azureCloudState.ClusterName + " " + azureCloudState.ResourceGroupName + " " + azureCloudState.Region
	default:
		return errors.New("[azure] Invalid operation for init state")

	}

	ctx = context.Background()

	// replace

	if err := obj.Client.InitClient(storage); err != nil {
		return err
	}

	obj.Client.SetRegion(obj.Region)
	//
	//err := obj.setRequiredENV_VAR(storage, ctx)
	//if err != nil {
	//	return err
	//}
	//cred, err := azidentity.NewDefaultAzureCredential(nil)
	//if err != nil {
	//	return err
	//}
	//obj.AzureTokenCred = cred

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	storage.Logger().Success("[azure] init cloud state")

	return nil
}

func ReturnAzureStruct(metadata resources.Metadata, ClientOption func() AzureGo) (*AzureProvider, error) {

	return &AzureProvider{
		ClusterName: metadata.ClusterName,
		Region:      metadata.Region,
		HACluster:   metadata.IsHA,
		//ResourceGroup: metadata.ClusterName + "-azure-ha-ksctl",
		Metadata: Metadata{
			K8sVersion: metadata.K8sVersion,
			K8sName:    metadata.K8sDistro,
		},
		Client: ClientOption(),
	}, nil
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
	if err := isValidVMSize(cloud, size); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}
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
func (obj *AzureProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}
		if azureCloudState.InfoControlPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		return len(azureCloudState.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.Metadata.NoCP = no
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}

		currLen := len(azureCloudState.InfoControlPlanes.Names)
		if currLen == 0 {
			azureCloudState.InfoControlPlanes.Names = make([]string, no)
			azureCloudState.InfoControlPlanes.Hostnames = make([]string, no) // as we don't need it now
			azureCloudState.InfoControlPlanes.PublicIPs = make([]string, no)
			azureCloudState.InfoControlPlanes.PrivateIPs = make([]string, no)
			azureCloudState.InfoControlPlanes.DiskNames = make([]string, no)
			azureCloudState.InfoControlPlanes.NetworkInterfaceNames = make([]string, no)
			azureCloudState.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			azureCloudState.InfoControlPlanes.PublicIPNames = make([]string, no)
			azureCloudState.InfoControlPlanes.PublicIPIDs = make([]string, no)
		}
		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *AzureProvider) NoOfDataStore(no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}
		if azureCloudState.InfoDatabase.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		return len(azureCloudState.InfoDatabase.Names), nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.Metadata.NoDS = no

		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}

		currLen := len(azureCloudState.InfoDatabase.Names)
		if currLen == 0 {
			azureCloudState.InfoDatabase.Names = make([]string, no)
			azureCloudState.InfoDatabase.Hostnames = make([]string, no) // TODO: remove it: as we don't need it now
			azureCloudState.InfoDatabase.PublicIPs = make([]string, no)
			azureCloudState.InfoDatabase.PrivateIPs = make([]string, no)
			azureCloudState.InfoDatabase.DiskNames = make([]string, no)
			azureCloudState.InfoDatabase.NetworkInterfaceNames = make([]string, no)
			azureCloudState.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			azureCloudState.InfoDatabase.PublicIPNames = make([]string, no)
			azureCloudState.InfoDatabase.PublicIPIDs = make([]string, no)
		}

		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of Datastore>= 1 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (obj *AzureProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}
		if azureCloudState.InfoWorkerPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		return len(azureCloudState.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.Metadata.NoWP = no
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}
		currLen := len(azureCloudState.InfoWorkerPlanes.Names)

		newLen := no

		if currLen == 0 {
			azureCloudState.InfoWorkerPlanes.Names = make([]string, no)
			azureCloudState.InfoWorkerPlanes.Hostnames = make([]string, no)
			azureCloudState.InfoWorkerPlanes.PublicIPs = make([]string, no)
			azureCloudState.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			azureCloudState.InfoWorkerPlanes.DiskNames = make([]string, no)
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames = make([]string, no)
			azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			azureCloudState.InfoWorkerPlanes.PublicIPNames = make([]string, no)
			azureCloudState.InfoWorkerPlanes.PublicIPIDs = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					azureCloudState.InfoWorkerPlanes.Names = append(azureCloudState.InfoWorkerPlanes.Names, "")
					azureCloudState.InfoWorkerPlanes.Hostnames = append(azureCloudState.InfoWorkerPlanes.Hostnames, "")
					azureCloudState.InfoWorkerPlanes.PublicIPs = append(azureCloudState.InfoWorkerPlanes.PublicIPs, "")
					azureCloudState.InfoWorkerPlanes.PrivateIPs = append(azureCloudState.InfoWorkerPlanes.PrivateIPs, "")
					azureCloudState.InfoWorkerPlanes.DiskNames = append(azureCloudState.InfoWorkerPlanes.DiskNames, "")
					azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames = append(azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames, "")
					azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = append(azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					azureCloudState.InfoWorkerPlanes.PublicIPNames = append(azureCloudState.InfoWorkerPlanes.PublicIPNames, "")
					azureCloudState.InfoWorkerPlanes.PublicIPIDs = append(azureCloudState.InfoWorkerPlanes.PublicIPIDs, "")
				}
			} else {
				// for downscaling
				azureCloudState.InfoWorkerPlanes.Names = azureCloudState.InfoWorkerPlanes.Names[:newLen]
				azureCloudState.InfoWorkerPlanes.Hostnames = azureCloudState.InfoWorkerPlanes.Hostnames[:newLen]
				azureCloudState.InfoWorkerPlanes.PublicIPs = azureCloudState.InfoWorkerPlanes.PublicIPs[:newLen]
				azureCloudState.InfoWorkerPlanes.PrivateIPs = azureCloudState.InfoWorkerPlanes.PrivateIPs[:newLen]
				azureCloudState.InfoWorkerPlanes.DiskNames = azureCloudState.InfoWorkerPlanes.DiskNames[:newLen]
				azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames = azureCloudState.InfoWorkerPlanes.NetworkInterfaceNames[:newLen]
				azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs = azureCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				azureCloudState.InfoWorkerPlanes.PublicIPNames = azureCloudState.InfoWorkerPlanes.PublicIPNames[:newLen]
				azureCloudState.InfoWorkerPlanes.PublicIPIDs = azureCloudState.InfoWorkerPlanes.PublicIPIDs[:newLen]
			}
		}

		if err := saveStateHelper(storage); err != nil {
			return -1, err
		}

		return -1, nil
	}
	return -1, fmt.Errorf("[azure] constrains for no of workplane >= 0")
}

func GetRAWClusterInfos(storage resources.StorageFactory) ([]cloud_control_res.AllClusterData, error) {
	var data []cloud_control_res.AllClusterData

	// first get all the directories of ha
	haFolders, err := storage.Path(generatePath(utils.CLUSTER_PATH, utils.CLUSTER_TYPE_HA)).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range haFolders {
		path := generatePath(utils.CLUSTER_PATH, utils.CLUSTER_TYPE_HA, haFolder[0]+" "+haFolder[1]+" "+haFolder[2], STATE_FILE_NAME)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, err
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, err
		}
		data = append(data,
			cloud_control_res.AllClusterData{
				Provider: utils.CLOUD_AZURE,
				Name:     haFolder[0],
				Region:   haFolder[2],
				Type:     utils.CLUSTER_TYPE_HA,

				NoWP: len(clusterState.InfoWorkerPlanes.Names),
				NoCP: len(clusterState.InfoControlPlanes.Names),
				NoDS: len(clusterState.InfoDatabase.Names),

				K8sDistro:  clusterState.KubernetesDistro,
				K8sVersion: clusterState.KubernetesVer,
			})
	}

	managedFolders, err := storage.Path(generatePath(utils.CLUSTER_PATH, "managed")).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range managedFolders {

		path := generatePath(utils.CLUSTER_PATH, "managed", haFolder[0]+" "+haFolder[1]+" "+haFolder[2], STATE_FILE_NAME)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, err
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, err
		}

		data = append(data,
			cloud_control_res.AllClusterData{
				Provider:   utils.CLOUD_AZURE,
				Name:       haFolder[0],
				Region:     haFolder[2],
				Type:       utils.CLUSTER_TYPE_MANG,
				K8sDistro:  clusterState.KubernetesDistro,
				K8sVersion: clusterState.KubernetesVer,
				NoMgt:      clusterState.NoManagedNodes,
			})
	}
	return data, nil
}

func isPresent(storage resources.StorageFactory) bool {
	_, err := storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName, STATE_FILE_NAME)).Load()
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (obj *AzureProvider) SwitchCluster(storage resources.StorageFactory) error {

	switch obj.HACluster {
	case true:
		obj.ResourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.ClusterName, utils.CLUSTER_TYPE_HA)
		clusterDirName = obj.ClusterName + " " + obj.ResourceGroup + " " + obj.Region
		clusterType = utils.CLUSTER_TYPE_HA
		if isPresent(storage) {
			printKubeconfig(storage, utils.OPERATION_STATE_CREATE)
			return nil
		}
	case false:
		obj.ResourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.ClusterName, utils.CLUSTER_TYPE_MANG)
		clusterDirName = obj.ClusterName + " " + obj.ResourceGroup + " " + obj.Region
		clusterType = utils.CLUSTER_TYPE_MANG
		if isPresent(storage) {
			printKubeconfig(storage, utils.OPERATION_STATE_CREATE)
			return nil
		}
	}
	return fmt.Errorf("[azure] Cluster not found")
}
