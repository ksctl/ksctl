package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/pkg/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
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

type metadata struct {
	resName string
	role    KsctlRole
	vmType  string
	public  bool

	// apps    string
	cni     string
	version string

	// these are used for managing the state and are the size of the arrays
	noCP int
	noWP int
	noDS int

	k8sName    KsctlKubernetes
	k8sVersion string
}

type AzureProvider struct {
	clusterName   string
	haCluster     bool
	resourceGroup string
	region        string
	sshPath       string
	metadata
	mxName   sync.Mutex
	mxRole   sync.Mutex
	mxVMType sync.Mutex
	mxState  sync.Mutex

	client AzureGo
}

var (
	azureCloudState *StateConfiguration

	clusterDirName string
	clusterType    KsctlClusterType // it stores the ha or managed

	ctx context.Context
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

// GetSecretTokens implements resources.CloudFactory.
func (*AzureProvider) GetSecretTokens(resources.StorageFactory) (map[string][]byte, error) {

	envTenant := os.Getenv("AZURE_TENANT_ID")
	envSub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	envClientid := os.Getenv("AZURE_CLIENT_ID")
	envClientsec := os.Getenv("AZURE_CLIENT_SECRET")

	return map[string][]byte{
		"AZURE_TENANT_ID":       []byte(envTenant),
		"AZURE_SUBSCRIPTION_ID": []byte(envSub),
		"AZURE_CLIENT_ID":       []byte(envClientid),
		"AZURE_CLIENT_SECRET":   []byte(envClientsec),
	}, nil
}

// GetStateFile implements resources.CloudFactory.
func (*AzureProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(azureCloudState)
	if err != nil {
		return "", err
	}
	return string(cloudstate), nil
}

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

	obj.metadata.k8sVersion = ver
	return obj
}

type Credential struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
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
		// public IPs
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
func (obj *AzureProvider) InitState(storage resources.StorageFactory, operation KsctlOperation) error {

	switch obj.haCluster {
	case false:
		clusterType = ClusterTypeMang
	case true:
		clusterType = ClusterTypeHa
	}
	obj.resourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.clusterName, clusterType)
	clusterDirName = obj.clusterName + " " + obj.resourceGroup + " " + obj.region

	errLoadState := loadStateHelper(storage)
	switch operation {
	case OperationStateCreate:
		if errLoadState == nil && azureCloudState.IsCompleted {
			return fmt.Errorf("[azure] already exist")
		}
		if errLoadState == nil && !azureCloudState.IsCompleted {
			storage.Logger().Note("[azure] RESUME triggered!!")
		} else {
			storage.Logger().Note("[azure] Fresh state!!")
			azureCloudState = &StateConfiguration{
				IsCompleted:      false,
				ClusterName:      obj.clusterName,
				Region:           obj.region,
				KubernetesDistro: string(obj.metadata.k8sName),
				KubernetesVer:    obj.metadata.k8sVersion,
			}
		}

	case OperationStateDelete:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[azure] Delete resource(s)")

	case OperationStateGet:
		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[azure] Get resources")
		clusterDirName = azureCloudState.ClusterName + " " + azureCloudState.ResourceGroupName + " " + azureCloudState.Region
	default:
		return errors.New("[azure] Invalid operation for init state")

	}

	ctx = context.Background()

	if err := obj.client.InitClient(storage); err != nil {
		return err
	}

	// added the resource grp and region for easy of use for the client library
	obj.client.SetRegion(obj.region)
	obj.client.SetResourceGrp(obj.resourceGroup)

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	storage.Logger().Success("[azure] init cloud state")

	return nil
}

func ReturnAzureStruct(meta resources.Metadata, ClientOption func() AzureGo) (*AzureProvider, error) {

	return &AzureProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
			k8sName:    meta.K8sDistro,
		},
		client: ClientOption(),
	}, nil
}

// Name it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) resources.CloudFactory {
	cloud.mxName.Lock()

	if err := utils.IsValidName(resName); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}
	cloud.metadata.resName = resName
	return cloud
}

// Role it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole KsctlRole) resources.CloudFactory {
	cloud.mxRole.Lock()

	switch resRole {
	case RoleCp, RoleDs, RoleLb, RoleWp:
		cloud.metadata.role = resRole
		return cloud
	default:
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err("invalid role assumed")
		return nil
	}
}

// VMType it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) resources.CloudFactory {
	cloud.mxVMType.Lock()

	cloud.metadata.vmType = size
	if err := isValidVMSize(cloud, size); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}
	return cloud
}

// Visibility whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.metadata.public = toBePublic
	return cloud
}

func (cloud *AzureProvider) Application(s string) (externalApps bool) {
	return true
}

// CNI Why will be installed because it will be done by the extensions
func (cloud *AzureProvider) CNI(s string) (externalCNI bool) {

	switch KsctlValidCNIPlugin(s) {
	case CNIKubenet, CNIAzure:
		cloud.metadata.cni = s
	case "":
		cloud.metadata.cni = string(CNIAzure)
	default:
		cloud.metadata.cni = string(CNINone) // any other cni it will marked as none for NetworkPlugin
		return true
	}

	return false
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
		obj.metadata.noCP = no
		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}

		currLen := len(azureCloudState.InfoControlPlanes.Names)
		if currLen == 0 {
			azureCloudState.InfoControlPlanes.Names = make([]string, no)
			azureCloudState.InfoControlPlanes.Hostnames = make([]string, no)
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
		obj.metadata.noDS = no

		if azureCloudState == nil {
			return -1, fmt.Errorf("[azure] state init not called")
		}

		currLen := len(azureCloudState.InfoDatabase.Names)
		if currLen == 0 {
			azureCloudState.InfoDatabase.Names = make([]string, no)
			azureCloudState.InfoDatabase.Hostnames = make([]string, no)
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
		obj.metadata.noWP = no
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
	haFolders, err := storage.Path(generatePath(UtilClusterPath, ClusterTypeHa)).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range haFolders {
		path := generatePath(UtilClusterPath, ClusterTypeHa, haFolder[0]+" "+haFolder[1]+" "+haFolder[2], STATE_FILE_NAME)
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
				Provider: CloudAzure,
				Name:     haFolder[0],
				Region:   haFolder[2],
				Type:     ClusterTypeHa,

				NoWP: len(clusterState.InfoWorkerPlanes.Names),
				NoCP: len(clusterState.InfoControlPlanes.Names),
				NoDS: len(clusterState.InfoDatabase.Names),

				K8sDistro:  KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
			})
	}

	managedFolders, err := storage.Path(generatePath(UtilClusterPath, ClusterTypeMang)).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range managedFolders {

		path := generatePath(UtilClusterPath, ClusterTypeMang, haFolder[0]+" "+haFolder[1]+" "+haFolder[2], STATE_FILE_NAME)
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
				Provider:   CloudAzure,
				Name:       haFolder[0],
				Region:     haFolder[2],
				Type:       ClusterTypeMang,
				K8sDistro:  KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
				NoMgt:      clusterState.NoManagedNodes,
			})
	}
	return data, nil
}

func isPresent(storage resources.StorageFactory) bool {
	_, err := storage.Path(utils.GetPath(UtilClusterPath, CloudAzure, clusterType, clusterDirName, STATE_FILE_NAME)).Load()
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (obj *AzureProvider) SwitchCluster(storage resources.StorageFactory) error {

	switch obj.haCluster {
	case true:
		obj.resourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.clusterName, ClusterTypeHa)
		clusterDirName = obj.clusterName + " " + obj.resourceGroup + " " + obj.region
		clusterType = ClusterTypeHa
		if isPresent(storage) {
			printKubeconfig(storage, OperationStateCreate)
			return nil
		}
	case false:
		obj.resourceGroup = fmt.Sprintf("%s-ksctl-%s-resgrp", obj.clusterName, ClusterTypeMang)
		clusterDirName = obj.clusterName + " " + obj.resourceGroup + " " + obj.region
		clusterType = ClusterTypeMang
		if isPresent(storage) {
			printKubeconfig(storage, OperationStateCreate)
			return nil
		}
	}
	return fmt.Errorf("[azure] Cluster not found")
}
