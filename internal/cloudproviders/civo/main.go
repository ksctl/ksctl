package civo

import (
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

type InstanceID struct {
	ControlNodes     []string `json:"controlnodeids"`
	WorkerNodes      []string `json:"workernodeids"`
	LoadBalancerNode string   `json:"loadbalancernodeid"`
	DatabaseNode     []string `json:"databasenodeids"`
}

type HostNames struct {
	ControlNodes     []string `json:"controlnode"`
	WorkerNodes      []string `json:"workernode"`
	LoadBalancerNode string   `json:"loadbalancernode"`
	DatabaseNode     []string `json:"databasenodes"`
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

	ClusterName      string `json:"clustername"`
	Region           string `json:"region"`
	ManagedClusterID string `json:"managed_cluster_id"`
	NoManagedNodes   int    `json:"no_managed_cluster_nodes"`

	SSHID            string `json:"ssh_id"`
	SSHUser          string `json:"ssh_usr"`
	SSHPrivateKeyLoc string `json:"ssh_private_key_location"`

	InstanceIDs InstanceID `json:"instanceids"`
	NetworkIDs  NetworkID  `json:"networkids"`
	IPv4        InstanceIP `json:"ipv4_addr"`
	HostNames   `json:"hostnames"`

	KubernetesDistro string `json:"k8s_distro"`
	KubernetesVer    string `json:"k8s_version"`
}

var (
	civoCloudState *StateConfiguration
	clusterDirName string
	clusterType    KsctlClusterType // it stores the ha or managed
)

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

type metadata struct {
	resName string
	role    KsctlRole
	vmType  string
	public  bool

	// purpose: application in managed cluster
	apps string
	cni  string
	// these are used for managing the state and are the size of the arrays
	noCP int
	noWP int
	noDS int

	k8sName    KsctlKubernetes
	k8sVersion string
}

func (m metadata) String() string {
	return fmt.Sprintf("{ resName: '%s', role: '%s', vmtype: '%s' }\n", m.resName, m.role, m.vmType)
}

type CivoProvider struct {
	clusterName string
	haCluster   bool
	region      string

	mxName   sync.Mutex
	mxRole   sync.Mutex
	mxVMType sync.Mutex
	mxState  sync.Mutex

	metadata

	client CivoGo
}

// GetSecretTokens implements resources.CloudFactory.
func (this *CivoProvider) GetSecretTokens(storage resources.StorageFactory) (map[string][]byte, error) {
	return map[string][]byte{
		"CIVO_TOKEN": []byte(fetchAPIKey(storage)), // use base64 conversion
	}, nil
}

// GetStateFile implements resources.CloudFactory.
func (*CivoProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(civoCloudState)
	if err != nil {
		return "", err
	}
	return string(cloudstate), nil
}

type Credential struct {
	Token string `json:"token"`
}

// GetStateForHACluster implements resources.CloudFactory.
// WARN: the array copy is a shallow copy
func (client *CivoProvider) GetStateForHACluster(storage resources.StorageFactory) (cloud_control_res.CloudResourceState, error) {

	payload := cloud_control_res.CloudResourceState{
		SSHState: cloud_control_res.SSHInfo{
			PathPrivateKey: civoCloudState.SSHPrivateKeyLoc,
			UserName:       civoCloudState.SSHUser,
		},
		Metadata: cloud_control_res.Metadata{
			ClusterName: client.clusterName,
			Provider:    CloudCivo,
			Region:      client.region,
			ClusterType: clusterType,
			ClusterDir:  clusterDirName,
		},
		// public IPs
		IPv4ControlPlanes: civoCloudState.IPv4.IPControlplane,
		IPv4DataStores:    civoCloudState.IPv4.IPDataStore,
		IPv4WorkerPlanes:  civoCloudState.IPv4.IPWorkerPlane,
		IPv4LoadBalancer:  civoCloudState.IPv4.IPLoadbalancer,

		// Private IPs
		PrivateIPv4ControlPlanes: civoCloudState.IPv4.PrivateIPControlplane,
		PrivateIPv4DataStores:    civoCloudState.IPv4.PrivateIPDataStore,
		PrivateIPv4LoadBalancer:  civoCloudState.IPv4.PrivateIPLoadbalancer,
	}
	storage.Logger().Success("[civo] Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *CivoProvider) InitState(storage resources.StorageFactory, operation KsctlOperation) error {

	clusterDirName = obj.clusterName + " " + obj.region
	if obj.haCluster {
		clusterType = ClusterTypeHa
	} else {
		clusterType = ClusterTypeMang
	}

	civoCloudState = &StateConfiguration{}
	errLoadState := loadStateHelper(storage, generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME))

	switch operation {
	case OperationStateCreate:
		if errLoadState == nil && civoCloudState.IsCompleted {
			// then found and it and the process is done then no point of duplicate creation
			return fmt.Errorf("[civo] already exist")
		}

		if errLoadState == nil && !civoCloudState.IsCompleted {
			// file present but not completed
			storage.Logger().Note("[civo] RESUME triggered!!")
		} else {
			storage.Logger().Note("[civo] Fresh state!!")
			civoCloudState = &StateConfiguration{
				IsCompleted:      false,
				Region:           obj.region,
				ClusterName:      obj.clusterName,
				KubernetesDistro: string(obj.k8sName),
				KubernetesVer:    obj.k8sVersion,
			}
		}

	case OperationStateGet:

		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[civo] Get resources")

	case OperationStateDelete:

		if errLoadState != nil {
			return fmt.Errorf("no cluster state found reason:%s\n", errLoadState.Error())
		}
		storage.Logger().Note("[civo] Delete resource(s)")
	default:
		return errors.New("[civo] Invalid operation for init state")
	}

	if err := obj.client.InitClient(storage, obj.region); err != nil {
		return err
	}

	if err := validationOfArguments(obj); err != nil {
		return err
	}
	storage.Logger().Success("[civo] init cloud state")
	return nil
}

func ReturnCivoStruct(meta resources.Metadata, ClientOption func() CivoGo) (*CivoProvider, error) {
	return &CivoProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sName:    meta.K8sDistro,
			k8sVersion: meta.K8sVersion,
		},
		client: ClientOption(),
	}, nil
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudFactory {
	cloud.mxName.Lock()

	if err := utils.IsValidName(resName); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return nil
	}
	cloud.metadata.resName = resName
	//fmt.Println("[trigger] Name", cloud.metadata)
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole KsctlRole) resources.CloudFactory {
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

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudFactory {
	cloud.mxVMType.Lock()

	if err := isValidVMSize(cloud, size); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())

		return nil
	}
	cloud.metadata.vmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *CivoProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.metadata.public = toBePublic
	return cloud
}

// if its ha its always false instead it tells whether the provider has support in their managed offerering
func (cloud *CivoProvider) SupportForApplications() bool {
	return true
}

func aggregratedApps(s string) (ret string) {
	if len(s) == 0 {
		ret = "traefik2-nodeport,metrics-server" // default: applications
	} else {
		ret = s + ",traefik2-nodeport,metrics-server"
	}
	return
}

func (client *CivoProvider) Application(s string) (externalApps bool) {
	client.metadata.apps = aggregratedApps(s)
	return false
}

func (client *CivoProvider) CNI(s string) (externalCNI bool) {

	switch KsctlValidCNIPlugin(s) {
	case CNICilium, CNIFlannel:
		client.metadata.cni = s
	case "":
		client.metadata.cni = string(CNIFlannel)
	default:
		// nothing external
		client.metadata.cni = string(CNINone)
		return true
	}

	return false
}

func k8sVersion(obj *CivoProvider, ver string) string {
	if len(ver) == 0 {
		return "1.26.4-k3s1"
	}

	ver = ver + "-k3s1"
	if err := isValidK8sVersion(obj, ver); err != nil {
		var logFactory logger.LogFactory = &logger.Logger{}
		logFactory.Err(err.Error())
		return ""
	}
	return ver
}

// Version implements resources.CloudFactory.
func (obj *CivoProvider) Version(ver string) resources.CloudFactory {
	obj.metadata.k8sVersion = k8sVersion(obj, ver)
	if len(obj.metadata.k8sVersion) == 0 {
		return nil
	}
	return obj
}

func (*CivoProvider) GetHostNameAllWorkerNode() []string {
	var hostnames []string = make([]string, len(civoCloudState.HostNames.WorkerNodes))
	copy(hostnames, civoCloudState.HostNames.WorkerNodes)
	return hostnames
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *CivoProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}
		if civoCloudState.InstanceIDs.ControlNodes == nil {
			return -1, fmt.Errorf("[civo] unable to fetch controlplane instanceID")
		}
		return len(civoCloudState.InstanceIDs.ControlNodes), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}

		currLen := len(civoCloudState.InstanceIDs.ControlNodes)
		if currLen == 0 {
			civoCloudState.InstanceIDs.ControlNodes = make([]string, no)
			civoCloudState.IPv4.IPControlplane = make([]string, no)
			civoCloudState.IPv4.PrivateIPControlplane = make([]string, no)
			civoCloudState.HostNames.ControlNodes = make([]string, no)
		}
		return -1, nil
	}
	return -1, fmt.Errorf("[civo] constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *CivoProvider) NoOfDataStore(no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}
		if civoCloudState.InstanceIDs.DatabaseNode == nil {
			return -1, fmt.Errorf("[civo] unable to fetch DataStore instanceID")
		}
		return len(civoCloudState.InstanceIDs.DatabaseNode), nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.metadata.noDS = no

		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}

		currLen := len(civoCloudState.InstanceIDs.DatabaseNode)
		if currLen == 0 {
			civoCloudState.InstanceIDs.DatabaseNode = make([]string, no)
			civoCloudState.IPv4.IPDataStore = make([]string, no)
			civoCloudState.IPv4.PrivateIPDataStore = make([]string, no)
			civoCloudState.HostNames.DatabaseNode = make([]string, no)
		}

		return -1, nil
	}
	return -1, fmt.Errorf("[civo] constrains for no of Datastore>= 1 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
// NOTE: make it better for wokerplane to save add stuff and remove stuff
func (obj *CivoProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}
		if civoCloudState.InstanceIDs.WorkerNodes == nil {
			return -1, fmt.Errorf("[civo] unable to fetch workerplane instanceID")
		}
		return len(civoCloudState.InstanceIDs.WorkerNodes), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if civoCloudState == nil {
			return -1, fmt.Errorf("[civo] state init not called!")
		}
		currLen := len(civoCloudState.InstanceIDs.WorkerNodes)

		newLen := no

		if currLen == 0 {
			civoCloudState.InstanceIDs.WorkerNodes = make([]string, no)
			civoCloudState.IPv4.IPWorkerPlane = make([]string, no)
			civoCloudState.IPv4.PrivateIPWorkerPlane = make([]string, no)
			civoCloudState.HostNames.WorkerNodes = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					civoCloudState.InstanceIDs.WorkerNodes = append(civoCloudState.InstanceIDs.WorkerNodes, "")
					civoCloudState.IPv4.IPWorkerPlane = append(civoCloudState.IPv4.IPWorkerPlane, "")
					civoCloudState.IPv4.PrivateIPWorkerPlane = append(civoCloudState.IPv4.PrivateIPWorkerPlane, "")
					civoCloudState.HostNames.WorkerNodes = append(civoCloudState.HostNames.WorkerNodes, "")
				}
			} else {
				// for downscaling
				civoCloudState.InstanceIDs.WorkerNodes = civoCloudState.InstanceIDs.WorkerNodes[:newLen]
				civoCloudState.IPv4.IPWorkerPlane = civoCloudState.IPv4.IPWorkerPlane[:newLen]
				civoCloudState.IPv4.PrivateIPWorkerPlane = civoCloudState.IPv4.PrivateIPWorkerPlane[:newLen]
				civoCloudState.HostNames.WorkerNodes = civoCloudState.HostNames.WorkerNodes[:newLen]
			}
		}
		path := generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

		if err := saveStateHelper(storage, path); err != nil {
			return -1, err
		}

		return -1, nil
	}
	return -1, fmt.Errorf("[civo] constrains for no of workplane >= 0")
}

func GetRAWClusterInfos(storage resources.StorageFactory) ([]cloud_control_res.AllClusterData, error) {
	var data []cloud_control_res.AllClusterData

	// first get all the directories of ha
	haFolders, err := storage.Path(generatePath(UtilClusterPath, ClusterTypeHa)).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range haFolders {
		path := generatePath(UtilClusterPath, ClusterTypeHa, haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
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
				Provider: CloudCivo,
				Name:     haFolder[0],
				Region:   haFolder[1],
				Type:     ClusterTypeHa,

				NoWP: len(clusterState.InstanceIDs.WorkerNodes),
				NoCP: len(clusterState.InstanceIDs.ControlNodes),
				NoDS: len(clusterState.InstanceIDs.DatabaseNode),

				K8sDistro:  KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
			})
	}

	managedFolders, err := storage.Path(generatePath(UtilClusterPath, "managed")).GetFolders()
	if err != nil {
		return nil, err
	}

	for _, haFolder := range managedFolders {

		path := generatePath(UtilClusterPath, "managed", haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
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
				Provider:   CloudCivo,
				Name:       haFolder[0],
				Region:     haFolder[1],
				Type:       ClusterTypeMang,
				K8sDistro:  KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
				NoMgt:      clusterState.NoManagedNodes,
			})
	}
	return data, nil
}

func isPresent(storage resources.StorageFactory) bool {
	_, err := storage.Path(utils.GetPath(UtilClusterPath, CloudCivo, clusterType, clusterDirName, STATE_FILE_NAME)).Load()
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (obj *CivoProvider) SwitchCluster(storage resources.StorageFactory) error {
	clusterDirName = obj.clusterName + " " + obj.region
	switch obj.haCluster {
	case true:
		clusterType = ClusterTypeHa
		if isPresent(storage) {
			printKubeconfig(storage, OperationStateCreate)
			return nil
		}
	case false:
		clusterType = ClusterTypeMang
		if isPresent(storage) {
			printKubeconfig(storage, OperationStateCreate)
			return nil
		}
	}
	return fmt.Errorf("[civo] Cluster not found")
}
