package civo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/logger"

	"github.com/civo/civogo"
	"github.com/kubesimplify/ksctl/api/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/utils"
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
	// at initial phase its building, only after the creation is done
	// it has "DONE" status otherwise "BUILDING"
	IsCompleted bool `json:"status"`

	ClusterName      string     `json:"clustername"`
	Region           string     `json:"region"`
	ManagedClusterID string     `json:"managed_cluster_id"`
	NoManagedNodes   int        `json:"no_managed_cluster_nodes"`
	SSHID            string     `json:"ssh_id"`
	InstanceIDs      InstanceID `json:"instanceids"`
	NetworkIDs       NetworkID  `json:"networkids"`
	IPv4             InstanceIP `json:"ipv4_addr"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}

var (
	civoCloudState *StateConfiguration
	civoClient     *civogo.Client
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

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool

	// purpose: application in managed cluster
	Apps    string
	Cni     string
	Version string
}

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`

	SSHPath string `json:"ssh_key"` // do check what need to be here

	Metadata

	// Application      string `json:"application"`
	// CNIPlugin        string `json:"cni_plugin"`
}

type Credential struct {
	Token string `json:"token"`
}

// GetStateForHACluster implements resources.CloudInfrastructure.
// TODO: add the steps to transfer data
func (client *CivoProvider) GetStateForHACluster(storage resources.StorageInfrastructure) (cloud_control_res.CloudResourceState, error) {
	payload := cloud_control_res.CloudResourceState{
		SSHState:          cloud_control_res.SSHPayload{PathPrivateKey: "abcd/rdcewcf"},
		Metadata:          cloud_control_res.Metadata{ClusterName: client.ClusterName},
		IPv4ControlPlanes: civoCloudState.InstanceIDs.ControlNodes,
	}
	storage.Logger().Success("Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *CivoProvider) InitState(storage resources.StorageInfrastructure, operation string) error {

	clusterDirName = obj.ClusterName + " " + obj.Region
	if obj.HACluster {
		clusterType = "ha"
	} else {
		clusterType = "managed"
	}

	var err error
	civoCloudState = &StateConfiguration{}
	errLoadState := loadStateHelper(storage, generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME))

	switch operation {
	case "create":
		if errLoadState == nil && civoCloudState.IsCompleted {
			// then found and it and the process is done then no point of duplicate creation
			return fmt.Errorf("already exist")
		}

		if errLoadState == nil && !civoCloudState.IsCompleted {
			// file present but not completed
			storage.Logger().Note("RESUME triggered!!")
		} else {
			storage.Logger().Note("Fresh state!!")
			civoCloudState = &StateConfiguration{
				IsCompleted: false,
				Region:      obj.Region,
				ClusterName: obj.ClusterName,
			}
		}

	case "delete":

		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("Delete resource(s)")
	default:
		return errors.New("Invalid operation for init state")
	}

	civoClient, err = civogo.NewClient(fetchAPIKey(storage), obj.Region)
	if err != nil {
		return err
	}

	if err := validationOfArguments(obj.ClusterName, obj.Region); err != nil {
		return err
	}
	storage.Logger().Success("[civo] init cloud state")
	return nil
}

func ReturnCivoStruct(metadata resources.Metadata) (*CivoProvider, error) {
	return &CivoProvider{
		ClusterName: metadata.ClusterName,
		Region:      metadata.Region,
		HACluster:   metadata.IsHA,
	}, nil
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudInfrastructure {
	if err := utils.IsValidName(resName); err != nil {
		return nil
	}
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole string) resources.CloudInfrastructure {
	switch resRole {
	case "controlplane", "workerplane", "loadbalancer", "datastore":
		cloud.Metadata.Role = resRole
		return cloud
	default:
		return nil
	}
}

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudInfrastructure {
	if err := isValidVMSize(size); err != nil {
		return nil
	}
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *CivoProvider) Visibility(toBePublic bool) resources.CloudInfrastructure {
	cloud.Metadata.Public = toBePublic
	return cloud
}

// if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *CivoProvider) SupportForApplications() bool {
	return true
}

func (cloud *CivoProvider) SupportForCNI() bool {
	return true
}

func (client *CivoProvider) Application(s string) resources.CloudInfrastructure {
	if len(s) == 0 {
		client.Metadata.Apps = "Traefik-v2-nodeport,metrics-server" // default: applications
	} else {
		client.Metadata.Apps += ",Traefik-v2-nodeport,metrics-server"
	}

	return client
}

func (client *CivoProvider) CNI(s string) resources.CloudInfrastructure {
	if len(s) == 0 {
		client.Metadata.Cni = "flannel"
	} else {
		client.Metadata.Cni = s
	}
	return client
}

// Version implements resources.CloudInfrastructure.
func (obj *CivoProvider) Version(ver string) resources.CloudInfrastructure {
	if len(ver) == 0 {
		obj.Metadata.Version = "1.26.4-k3s1"
	} else {
		ver = ver + "-k3s1"
		if err := isValidK8sVersion(ver); err != nil {
			var logFactory logger.LogFactory = &logger.Logger{}
			logFactory.Err(err.Error())
			return nil
		}
		obj.Metadata.Version = ver
	}
	return obj
}

func GetRAWClusterInfos(storage resources.StorageInfrastructure) ([]cloud_control_res.AllClusterData, error) {
	var data []cloud_control_res.AllClusterData

	// first get all the directories of ha
	haFolders, err := storage.Path(generatePath(utils.CLUSTER_PATH, "ha")).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range haFolders {
		path := generatePath(utils.CLUSTER_PATH, "ha", haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
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
				Provider: "civo",
				Name:     haFolder[0],
				Region:   haFolder[1],
				Type:     "ha",

				NoWP: len(clusterState.InstanceIDs.WorkerNodes),
				NoCP: len(clusterState.InstanceIDs.ControlNodes),
				NoDS: len(clusterState.InstanceIDs.DatabaseNode),

				K8sDistro:  clusterState.KubernetesDistro,
				K8sVersion: clusterState.KubernetesVer,
			})
		// to fetch more info we need to read the state files
	}

	managedFolders, err := storage.Path(generatePath(utils.CLUSTER_PATH, "managed")).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range managedFolders {

		path := generatePath(utils.CLUSTER_PATH, "managed", haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
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
				Provider:   "civo",
				Name:       haFolder[0],
				Region:     haFolder[1],
				Type:       "managed",
				K8sDistro:  clusterState.KubernetesDistro,
				K8sVersion: clusterState.KubernetesVer,
				NoMgt:      clusterState.NoManagedNodes,
			})
	}
	return data, nil
}
