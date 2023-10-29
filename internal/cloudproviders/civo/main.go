package civo

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/pkg/resources"
	cloud_control_res "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
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
	clusterType    consts.KsctlClusterType // it stores the ha or managed

	log resources.LoggerFactory
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
	role    consts.KsctlRole
	vmType  string
	public  bool

	// purpose: application in managed cluster
	apps string
	cni  string
	// these are used for managing the state and are the size of the arrays
	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
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
		return "", log.NewError(err.Error())
	}

	log.Debug("Printing", "cloudstate", string(cloudstate))
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
			Provider:    consts.CloudCivo,
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
	log.Debug("Printing", "cloudState", payload)
	log.Success("Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *CivoProvider) InitState(storage resources.StorageFactory, operation consts.KsctlOperation) error {

	clusterDirName = obj.clusterName + " " + obj.region
	if obj.haCluster {
		clusterType = consts.ClusterTypeHa
	} else {
		clusterType = consts.ClusterTypeMang
	}

	civoCloudState = &StateConfiguration{}
	errLoadState := loadStateHelper(storage, generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME))

	switch operation {
	case consts.OperationStateCreate:
		if errLoadState == nil && civoCloudState.IsCompleted {
			// then found and it and the process is done then no point of duplicate creation
			return log.NewError("already exist")
		}

		if errLoadState == nil && !civoCloudState.IsCompleted {
			// file present but not completed
			log.Note("RESUME triggered!!")
		} else {
			log.Note("Fresh state!!")
			civoCloudState = &StateConfiguration{
				IsCompleted:      false,
				Region:           obj.region,
				ClusterName:      obj.clusterName,
				KubernetesDistro: string(obj.k8sName),
				KubernetesVer:    obj.k8sVersion,
			}
		}

	case consts.OperationStateGet:

		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Note("Get resources")

	case consts.OperationStateDelete:

		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Note("Delete resource(s)")
	default:
		return log.NewError("Invalid operation for init state")
	}

	if err := obj.client.InitClient(storage, obj.region); err != nil {
		return log.NewError(err.Error())
	}

	if err := validationOfArguments(obj); err != nil {
		return log.NewError(err.Error())
	}
	log.Debug("Printing", "CivoProvider", obj)
	log.Success("init cloud state")
	return nil
}

func ReturnCivoStruct(meta resources.Metadata, ClientOption func() CivoGo) (*CivoProvider, error) {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudCivo))

	obj := &CivoProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sName:    meta.K8sDistro,
			k8sVersion: meta.K8sVersion,
		},
		client: ClientOption(),
	}
	log.Debug("Printing", "CivoProvider", obj)
	return obj, nil
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudFactory {
	cloud.mxName.Lock()

	if err := utils.IsValidName(resName); err != nil {
		log.Error(err.Error())
		return nil
	}
	cloud.metadata.resName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole consts.KsctlRole) resources.CloudFactory {
	cloud.mxRole.Lock()

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		cloud.metadata.role = resRole
		log.Debug("Printing", "Role", resRole)
		return cloud
	default:
		log.Error("invalid role assumed")
		return nil
	}
}

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudFactory {
	cloud.mxVMType.Lock()

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error(err.Error())
		return nil
	}
	cloud.metadata.vmType = size
	log.Debug("Printing", "VMSize", size)
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *CivoProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.metadata.public = toBePublic
	log.Debug("Printing", "willBePublic", toBePublic)
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
	log.Debug("Printing", "apps", ret)
	return
}

func (client *CivoProvider) Application(s string) (externalApps bool) {
	client.metadata.apps = aggregratedApps(s)
	return false
}

func (client *CivoProvider) CNI(s string) (externalCNI bool) {

	log.Debug("Printing", "cni", s)
	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNICilium, consts.CNIFlannel:
		client.metadata.cni = s
	case "":
		client.metadata.cni = string(consts.CNIFlannel)
	default:
		// nothing external
		client.metadata.cni = string(consts.CNINone)
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
		log.Error(err.Error())
		return ""
	}
	log.Debug("Printing", "k8sVersion", ver)
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
	log.Debug("Printing", "hostnameOfWorkerPlanes", hostnames)
	return hostnames
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *CivoProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
		}
		if civoCloudState.InstanceIDs.ControlNodes == nil {
			return -1, log.NewError("unable to fetch controlplane instanceID")
		}
		log.Debug("Printing", "InstanceIDsOfControlplanes", civoCloudState.InstanceIDs.ControlNodes)
		return len(civoCloudState.InstanceIDs.ControlNodes), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
		}

		currLen := len(civoCloudState.InstanceIDs.ControlNodes)
		if currLen == 0 {
			civoCloudState.InstanceIDs.ControlNodes = make([]string, no)
			civoCloudState.IPv4.IPControlplane = make([]string, no)
			civoCloudState.IPv4.PrivateIPControlplane = make([]string, no)
			civoCloudState.HostNames.ControlNodes = make([]string, no)
		}
		log.Debug("Printing", "civoCloudState.InstanceIDs.ControlNodes", civoCloudState.InstanceIDs.ControlNodes)
		log.Debug("Printing", "civoCloudState.IPv4.IPControlplane", civoCloudState.IPv4.IPControlplane)
		log.Debug("Printing", "civoCloudState.IPv4.PrivateIPControlplane", civoCloudState.IPv4.PrivateIPControlplane)
		log.Debug("Printing", "civoCloudState.HostNames.ControlNodes", civoCloudState.HostNames.ControlNodes)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *CivoProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
		}
		if civoCloudState.InstanceIDs.DatabaseNode == nil {
			return -1, log.NewError("unable to fetch DataStore instanceID")
		}

		log.Debug("Printing", "InstanceIDsOfDatabaseNode", civoCloudState.InstanceIDs.DatabaseNode)

		return len(civoCloudState.InstanceIDs.DatabaseNode), nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.metadata.noDS = no

		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
		}

		currLen := len(civoCloudState.InstanceIDs.DatabaseNode)
		if currLen == 0 {
			civoCloudState.InstanceIDs.DatabaseNode = make([]string, no)
			civoCloudState.IPv4.IPDataStore = make([]string, no)
			civoCloudState.IPv4.PrivateIPDataStore = make([]string, no)
			civoCloudState.HostNames.DatabaseNode = make([]string, no)
		}

		log.Debug("Printing", "civoCloudState.InstanceIDs.DatabaseNode", civoCloudState.InstanceIDs.DatabaseNode)
		log.Debug("Printing", "civoCloudState.IPv4.IPDataStore", civoCloudState.IPv4.IPDataStore)
		log.Debug("Printing", "civoCloudState.IPv4.PrivateIPDataStore", civoCloudState.IPv4.PrivateIPDataStore)
		log.Debug("Printing", "civoCloudState.HostNames.DatabaseNode", civoCloudState.HostNames.DatabaseNode)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 1 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
// NOTE: make it better for wokerplane to save add stuff and remove stuff
func (obj *CivoProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
		}
		if civoCloudState.InstanceIDs.WorkerNodes == nil {
			return -1, log.NewError("unable to fetch workerplane instanceID")
		}

		log.Debug("Printing", "InstanceIDsOfWorkerPlane", civoCloudState.InstanceIDs.WorkerNodes)

		return len(civoCloudState.InstanceIDs.WorkerNodes), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if civoCloudState == nil {
			return -1, log.NewError("state init not called!")
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
		path := generatePath(consts.UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

		if err := saveStateHelper(storage, path); err != nil {
			return -1, err
		}

		log.Debug("Printing", "civoCloudState.InstanceIDs.WorkerNodes", civoCloudState.InstanceIDs.WorkerNodes)
		log.Debug("Printing", "civoCloudState.IPv4.IPWorkerPlane", civoCloudState.IPv4.IPWorkerPlane)
		log.Debug("Printing", "civoCloudState.IPv4.PrivateIPWorkerPlane", civoCloudState.IPv4.PrivateIPWorkerPlane)
		log.Debug("Printing", "civoCloudState.HostNames.WorkerNodes", civoCloudState.HostNames.WorkerNodes)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of workerplane >= 0")
}

func GetRAWClusterInfos(storage resources.StorageFactory) ([]cloud_control_res.AllClusterData, error) {
	var data []cloud_control_res.AllClusterData

	// first get all the directories of ha
	haFolders, err := storage.Path(generatePath(consts.UtilClusterPath, consts.ClusterTypeHa)).GetFolders()
	if err != nil {
		return nil, log.NewError(err.Error())
	}

	for _, haFolder := range haFolders {
		path := generatePath(consts.UtilClusterPath, consts.ClusterTypeHa, haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, log.NewError(err.Error())
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, log.NewError(err.Error())
		}
		data = append(data,
			cloud_control_res.AllClusterData{
				Provider: consts.CloudCivo,
				Name:     haFolder[0],
				Region:   haFolder[1],
				Type:     consts.ClusterTypeHa,

				NoWP: len(clusterState.InstanceIDs.WorkerNodes),
				NoCP: len(clusterState.InstanceIDs.ControlNodes),
				NoDS: len(clusterState.InstanceIDs.DatabaseNode),

				K8sDistro:  consts.KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
			})
		log.Debug("Printing", "cloudClusterInfoFeteched", data)
	}

	managedFolders, err := storage.Path(generatePath(consts.UtilClusterPath, "managed")).GetFolders()
	if err != nil {
		return nil, log.NewError(err.Error())
	}

	for _, haFolder := range managedFolders {

		path := generatePath(consts.UtilClusterPath, "managed", haFolder[0]+" "+haFolder[1], STATE_FILE_NAME)
		raw, err := storage.Path(path).Load()
		if err != nil {
			return nil, log.NewError(err.Error())
		}
		var clusterState *StateConfiguration
		if err := json.Unmarshal(raw, &clusterState); err != nil {
			return nil, log.NewError(err.Error())
		}

		data = append(data,
			cloud_control_res.AllClusterData{
				Provider:   consts.CloudCivo,
				Name:       haFolder[0],
				Region:     haFolder[1],
				Type:       consts.ClusterTypeMang,
				K8sDistro:  consts.KsctlKubernetes(clusterState.KubernetesDistro),
				K8sVersion: clusterState.KubernetesVer,
				NoMgt:      clusterState.NoManagedNodes,
			})

		log.Debug("Printing", "cloudClusterInfoFetched", data)
	}
	return data, nil
}

func isPresent(storage resources.StorageFactory) bool {
	_, err := storage.Path(utils.GetPath(consts.UtilClusterPath, consts.CloudCivo, clusterType, clusterDirName, STATE_FILE_NAME)).Load()
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (obj *CivoProvider) SwitchCluster(storage resources.StorageFactory) error {
	clusterDirName = obj.clusterName + " " + obj.region
	switch obj.haCluster {
	case true:
		clusterType = consts.ClusterTypeHa
		if isPresent(storage) {
			printKubeconfig(storage, consts.OperationStateCreate)
			return nil
		}
	case false:
		clusterType = consts.ClusterTypeMang
		if isPresent(storage) {
			printKubeconfig(storage, consts.OperationStateCreate)
			return nil
		}
	}
	return log.NewError("Cluster not found")
}
