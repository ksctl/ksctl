package azure

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"

	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudcontrolres "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

type metadata struct {
	public bool

	cni     string
	version string

	// these are used for managing the state and are the size of the arrays
	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
	k8sVersion string
}

type AzureProvider struct {
	clusterName   string
	haCluster     bool
	resourceGroup string
	region        string
	sshPath       string
	metadata

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	mu sync.Mutex

	client AzureGo
}

var (
	mainStateDocument *types.StorageDocument
	clusterType       consts.KsctlClusterType // it stores the ha or managed
	ctx               context.Context
	log               resources.LoggerFactory
)

func (*AzureProvider) GetStateFile(resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", err
	}
	log.Debug("Printing", "cloudstate", cloudstate)
	return string(cloudstate), nil
}

func (*AzureProvider) GetHostNameAllWorkerNode() []string {
	var hostnames []string = make([]string, len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames))
	copy(hostnames, mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames)
	log.Debug("Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

// Version implements resources.CloudFactory.
func (obj *AzureProvider) Version(ver string) resources.CloudFactory {
	log.Debug("Printing", "K8sVersion", ver)
	if err := isValidK8sVersion(obj, ver); err != nil {
		log.Error(err.Error())
		return nil
	}

	obj.metadata.k8sVersion = ver
	return obj
}

// GetStateForHACluster implements resources.CloudFactory.
// WARN: the array copy is a shallow copy
func (*AzureProvider) GetStateForHACluster(storage resources.StorageFactory) (cloudcontrolres.CloudResourceState, error) {
	payload := cloudcontrolres.CloudResourceState{
		SSHState: cloudcontrolres.SSHInfo{
			PrivateKey: mainStateDocument.SSHKeyPair.PrivateKey,
			UserName:   mainStateDocument.CloudInfra.Azure.B.SSHUser,
		},
		Metadata: cloudcontrolres.Metadata{
			ClusterName: mainStateDocument.ClusterName,
			Provider:    mainStateDocument.InfraProvider,
			Region:      mainStateDocument.Region,
			ClusterType: clusterType,
		},
		// public IPs
		IPv4ControlPlanes: mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs,
		IPv4DataStores:    mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs,
		IPv4WorkerPlanes:  mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs,
		IPv4LoadBalancer:  mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs,
		PrivateIPv4DataStores:    mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs,
		PrivateIPv4LoadBalancer:  mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP,
	}
	log.Debug("Printing", "azureStateTransferPayload", payload)

	log.Success("Transferred Data, it's ready to be shipped!")
	return payload, nil
}

// InitState implements resources.CloudFactory.
func (obj *AzureProvider) InitState(storage resources.StorageFactory, operation consts.KsctlOperation) error {

	switch obj.haCluster {
	case false:
		clusterType = consts.ClusterTypeMang
	case true:
		clusterType = consts.ClusterTypeHa
	}

	obj.chResName = make(chan string, 1)
	obj.chRole = make(chan consts.KsctlRole, 1)
	obj.chVMType = make(chan string, 1)

	obj.resourceGroup = GenerateResourceGroupName(obj.clusterName, string(clusterType))

	errLoadState := loadStateHelper(storage)
	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && mainStateDocument.CloudInfra.Azure.B.IsCompleted {
			return log.NewError("cluster already exist")
		}
		if errLoadState == nil && !mainStateDocument.CloudInfra.Azure.B.IsCompleted {
			log.Debug("RESUME triggered!!")
		} else {
			log.Debug("Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudAzure
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.Region = obj.region
			mainStateDocument.CloudInfra = &types.InfrastructureState{
				Azure: &types.StateConfigurationAzure{},
			}
			mainStateDocument.CloudInfra.Azure.B.KubernetesVer = obj.metadata.k8sVersion
			mainStateDocument.CloudInfra.Azure.B.KubernetesDistro = string(obj.metadata.k8sName)
		}

	case consts.OperationDelete:
		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Delete resource(s)")

	case consts.OperationGet:
		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Get resources")
	default:
		return log.NewError("Invalid operation for init state")
	}

	ctx = context.Background()

	if err := obj.client.InitClient(storage); err != nil {
		return log.NewError(err.Error())
	}

	// added the resource grp and region for easy of use for the client library
	obj.client.SetRegion(obj.region)
	obj.client.SetResourceGrp(obj.resourceGroup)

	if err := validationOfArguments(obj); err != nil {
		return log.NewError(err.Error())
	}

	log.Debug("init cloud state")

	return nil
}

func ReturnAzureStruct(meta resources.Metadata, state *types.StorageDocument, ClientOption func() AzureGo) (*AzureProvider, error) {

	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAzure))

	mainStateDocument = state

	obj := &AzureProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
			k8sName:    meta.K8sDistro,
		},
		client: ClientOption(),
	}

	log.Debug("Printing", "AzureProvider", obj)

	return obj, nil
}

// Name it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) resources.CloudFactory {

	if err := helpers.IsValidName(resName); err != nil {
		log.Error(err.Error())
		return nil
	}

	cloud.chResName <- resName
	return cloud
}

// Role it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole consts.KsctlRole) resources.CloudFactory {

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		cloud.chRole <- resRole
		return cloud
	default:
		log.Error("invalid role assumed")

		return nil
	}
}

// VMType it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) resources.CloudFactory {

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error(err.Error())
		return nil
	}
	cloud.chVMType <- size

	return cloud
}

// Visibility whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) resources.CloudFactory {
	cloud.metadata.public = toBePublic
	return cloud
}

func (cloud *AzureProvider) Application(s []string) (externalApps bool) {
	return true
}

// CNI Why will be installed because it will be done by the extensions
func (cloud *AzureProvider) CNI(s string) (externalCNI bool) {

	log.Debug("Printing", "cni", s)

	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNIKubenet, consts.CNIAzure:
		cloud.metadata.cni = s
	case "":
		cloud.metadata.cni = string(consts.CNIAzure)
	default:
		cloud.metadata.cni = string(consts.CNINone) // any other cni it will marked as none for NetworkPlugin
		return true
	}

	return false
}

// NoOfControlPlane implements resources.CloudFactory.
func (obj *AzureProvider) NoOfControlPlane(no int, setter bool) (int, error) {

	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs = make([]string, no)
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes", mainStateDocument.CloudInfra.Azure.InfoControlPlanes)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements resources.CloudFactory.
func (obj *AzureProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoDatabase.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase.Names", mainStateDocument.CloudInfra.Azure.InfoDatabase.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoDatabase.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Azure.InfoDatabase.Names)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Azure.InfoDatabase.Names = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPIDs = make([]string, no)
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase", mainStateDocument.CloudInfra.Azure.InfoDatabase)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 3 and odd number")
}

// NoOfWorkerPlane implements resources.CloudFactory.
func (obj *AzureProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		log.Debug("Prnting", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		currLen := len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = make([]string, no)
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames, "")
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs, "")
				}
			} else {
				// for downscaling
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[:newLen]
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[:newLen]
			}
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return -1, err
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes)

		return -1, nil
	}
	return -1, log.NewError("constrains for no of workplane >= 0")
}

func GetRAWClusterInfos(storage resources.StorageFactory, meta resources.Metadata) ([]cloudcontrolres.AllClusterData, error) {

	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAzure))

	var data []cloudcontrolres.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAzure),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudcontrolres.AllClusterData{
				Provider: consts.CloudAzure,
				Name:     v.ClusterName,
				Region:   v.Region,
				Type:     K,

				NoWP:  len(v.CloudInfra.Azure.InfoWorkerPlanes.Names),
				NoCP:  len(v.CloudInfra.Azure.InfoControlPlanes.Names),
				NoDS:  len(v.CloudInfra.Azure.InfoDatabase.Names),
				NoMgt: v.CloudInfra.Azure.NoManagedNodes,

				K8sDistro:  consts.KsctlKubernetes(v.CloudInfra.Azure.B.KubernetesDistro),
				K8sVersion: v.CloudInfra.Azure.B.KubernetesVer,
			})
			log.Debug("Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func isPresent(storage resources.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) bool {
	err := storage.AlreadyCreated(consts.CloudAzure, region, name, ksctlClusterType)
	if err != nil {
		return false
	}
	return true
}

func (obj *AzureProvider) IsPresent(storage resources.StorageFactory) error {

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
