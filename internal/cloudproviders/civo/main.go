package civo

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/internal/storage/types"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	cloud_control_res "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	mainStateDocument *types.StorageDocument
	clusterType       consts.KsctlClusterType // it stores the ha or managed
	civoCtx           context.Context
	log               resources.LoggerFactory
)

type metadata struct {
	public bool

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

type CivoProvider struct {
	clusterName string
	haCluster   bool
	region      string

	mu sync.Mutex

	metadata

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client CivoGo
}

func (*CivoProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", log.NewError(err.Error())
	}

	log.Debug("Printing", "cloudstate", string(cloudstate))
	return string(cloudstate), nil
}

// GetStateForHACluster implements resources.CloudFactory.
func (client *CivoProvider) GetStateForHACluster(storage resources.StorageFactory) (cloud_control_res.CloudResourceState, error) {

	payload := cloud_control_res.CloudResourceState{
		SSHState: cloud_control_res.SSHInfo{
			PrivateKey: mainStateDocument.SSHKeyPair.PrivateKey,
			UserName:   mainStateDocument.CloudInfra.Civo.B.SSHUser,
		},
		Metadata: cloud_control_res.Metadata{
			ClusterName: client.clusterName,
			Provider:    consts.CloudCivo,
			Region:      client.region,
			ClusterType: clusterType,
		},
		// public IPs
		IPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  mainStateDocument.CloudInfra.Civo.InfoLoadBalancer.PrivateIP,
	}
	log.Debug("Printing", "cloudState", payload)
	log.Success("Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *CivoProvider) InitState(storage resources.StorageFactory, operation consts.KsctlOperation) error {

	if obj.haCluster {
		clusterType = consts.ClusterTypeHa
	} else {
		clusterType = consts.ClusterTypeMang
	}

	obj.chResName = make(chan string, 1)
	obj.chRole = make(chan consts.KsctlRole, 1)
	obj.chVMType = make(chan string, 1)

	errLoadState := loadStateHelper(storage)

	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && mainStateDocument.CloudInfra.Civo.B.IsCompleted {
			// then found and it and the process is done then no point of duplicate creation
			return log.NewError("already exist")
		}

		if errLoadState == nil && !mainStateDocument.CloudInfra.Civo.B.IsCompleted {
			// file present but not completed
			log.Debug("RESUME triggered!!")
		} else {
			log.Debug("Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudCivo
			mainStateDocument.Region = obj.region
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.CloudInfra = &types.InfrastructureState{
				Civo: &types.StateConfigurationCivo{},
			}
			mainStateDocument.CloudInfra.Civo.B.KubernetesVer = obj.k8sVersion
			mainStateDocument.CloudInfra.Civo.B.KubernetesDistro = string(obj.k8sName)
		}

	case consts.OperationGet:

		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Get resources")

	case consts.OperationDelete:

		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Delete resource(s)")
	default:
		return log.NewError("Invalid operation for init state")
	}

	if err := obj.client.InitClient(storage, obj.region); err != nil {
		return log.NewError(err.Error())
	}

	if err := validationOfArguments(obj); err != nil {
		return log.NewError(err.Error())
	}
	log.Debug("init cloud state")
	return nil
}

func (cloud *CivoProvider) Credential(storage resources.StorageFactory) error {

	log.Print(civoCtx, "Enter CIVO TOKEN")
	token, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}
	client, err := civogo.NewClient(token, "LON1")
	if err != nil {
		return err
	}
	id := client.GetAccountID()

	if len(id) == 0 {
		return log.NewError(civoCtx, "Invalid user")
	}
	log.Print(civoCtx, "Recieved accountId", "userId", id)

	if err := storage.WriteCredentials(consts.CloudCivo,
		&types.CredentialsDocument{
			InfraProvider: consts.CloudCivo,
			Civo:          &types.CredentialsCivo{Token: token},
		}); err != nil {
		return err
	}

	return nil
}

func NewClient(parentCtx context.Context, meta resources.Metadata, parentLogger resources.LoggerFactory, state *types.StorageDocument, ClientOption func() CivoGo) (*CivoProvider, error) {
	log = parentLogger // intentional shallow copy so that we can use the same
	// logger to be used multiple places
	civoCtx = context.WithValue(parentCtx, consts.ContextModuleNameKey, string(consts.CloudCivo))

	mainStateDocument = state

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
	log.Debug(civoCtx, "Printing", "CivoProvider", obj)
	return obj, nil
}

// it will contain the name of the resource to be created
func (cloud *CivoProvider) Name(resName string) resources.CloudFactory {

	if err := helpers.IsValidName(resName); err != nil {
		log.Error(err.Error())
		return nil
	}
	cloud.chResName <- resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *CivoProvider) Role(resRole consts.KsctlRole) resources.CloudFactory {

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		cloud.chRole <- resRole
		log.Debug("Printing", "Role", resRole)
		return cloud
	default:
		log.Error("invalid role assumed")
		return nil
	}
}

// it will contain which vmType to create
func (cloud *CivoProvider) VMType(size string) resources.CloudFactory {

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error(err.Error())
		return nil
	}
	cloud.chVMType <- size
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

func aggregratedApps(s []string) (ret string) {
	if len(s) == 0 {
		ret = "traefik2-nodeport,metrics-server" // default: applications
	} else {
		ret = strings.Join(s, ",") + ",traefik2-nodeport,metrics-server"
	}
	log.Debug("Printing", "apps", ret)
	return
}

func (client *CivoProvider) Application(s []string) (externalApps bool) {
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
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
	log.Debug("Printing", "hostnameOfWorkerPlanes", hostnames)
	return hostnames
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *CivoProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}
		if mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs == nil {
			return -1, log.NewError("unable to fetch controlplane instanceID")
		}
		log.Debug("Printing", "InstanceIDsOfControlplanes", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		return len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}

		currLen := len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames = make([]string, no)
		}
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *CivoProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}
		if mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs == nil {
			return -1, log.NewError("unable to fetch DataStore instanceID")
		}

		log.Debug("Printing", "InstanceIDsOfDatabaseNode", mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)

		return len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}

		currLen := len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames = make([]string, no)
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs", mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames", mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 3 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
// NOTE: make it better for wokerplane to save add stuff and remove stuff
func (obj *CivoProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}
		if mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs == nil {
			return -1, log.NewError("unable to fetch workerplane instanceID")
		}

		log.Debug("Printing", "InstanceIDsOfWorkerPlane", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		return len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called!")
		}
		currLen := len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = append(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs, "")
					mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = append(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs, "")
					mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = append(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs, "")
					mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = append(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames, "")
				}
			} else {
				// for downscaling
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[:newLen]
			}
		}
		err := storage.Write(mainStateDocument)
		if err != nil {
			return -1, err
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs)
		log.Debug("Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of workerplane >= 0")
}

func GetRAWClusterInfos(storage resources.StorageFactory, meta resources.Metadata) ([]cloud_control_res.AllClusterData, error) {
	log = logger.NewStructuredLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudCivo))

	var data []cloud_control_res.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudCivo),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloud_control_res.AllClusterData{
				Provider: consts.CloudCivo,
				Name:     v.ClusterName,
				Region:   v.Region,
				Type:     K,

				NoWP:  len(v.CloudInfra.Civo.InfoWorkerPlanes.VMIDs),
				NoCP:  len(v.CloudInfra.Civo.InfoControlPlanes.VMIDs),
				NoDS:  len(v.CloudInfra.Civo.InfoDatabase.VMIDs),
				NoMgt: v.CloudInfra.Civo.NoManagedNodes,

				K8sDistro:  consts.KsctlKubernetes(v.CloudInfra.Civo.B.KubernetesDistro),
				K8sVersion: v.CloudInfra.Civo.B.KubernetesVer,
			})
			log.Debug("Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func isPresent(storage resources.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) bool {
	err := storage.AlreadyCreated(consts.CloudCivo, region, name, ksctlClusterType)
	return err == nil
}

func (obj *CivoProvider) IsPresent(storage resources.StorageFactory) error {
	switch obj.haCluster {
	case true:
		clusterType = consts.ClusterTypeHa
		if isPresent(storage, clusterType, obj.clusterName, obj.region) {
			return nil
		}
	case false:
		clusterType = consts.ClusterTypeMang
		if isPresent(storage, clusterType, obj.clusterName, obj.region) {
			return nil
		}
	}
	return log.NewError("Cluster not found")
}
