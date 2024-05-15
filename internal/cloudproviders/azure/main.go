package azure

import (
	"context"
	"encoding/json"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
	cloudcontrolres "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

type metadata struct {
	public bool

	cni string

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
	metadata

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	mu sync.Mutex

	client AzureGo
}

var (
	mainStateDocument *storageTypes.StorageDocument
	clusterType       consts.KsctlClusterType // it stores the ha or managed
	azureCtx          context.Context
	log               types.LoggerFactory
)

func (*AzureProvider) GetStateFile(types.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", err
	}
	log.Debug(azureCtx, "Printing", "cloudstate", cloudstate)
	return string(cloudstate), nil
}

func (*AzureProvider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames)
	log.Debug(azureCtx, "Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (obj *AzureProvider) Version(ver string) types.CloudFactory {
	log.Debug(azureCtx, "Printing", "K8sVersion", ver)
	if err := isValidK8sVersion(obj, ver); err != nil {
		log.Error(azureCtx, "azure.Version()", "err", err.Error())
		return nil
	}

	obj.metadata.k8sVersion = ver
	return obj
}

func (*AzureProvider) GetStateForHACluster(storage types.StorageFactory) (cloudcontrolres.CloudResourceState, error) {
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
		IPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  mainStateDocument.CloudInfra.Azure.InfoLoadBalancer.PrivateIP,
	}
	log.Debug(azureCtx, "Printing", "azureStateTransferPayload", payload)

	log.Success(azureCtx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
}

// InitState implements types.CloudFactory.
func (obj *AzureProvider) InitState(storage types.StorageFactory, operation consts.KsctlOperation) error {

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
			return log.NewError(azureCtx, "cluster already exist")
		}
		if errLoadState == nil && !mainStateDocument.CloudInfra.Azure.B.IsCompleted {
			log.Debug(azureCtx, "RESUME triggered!!")
		} else {
			log.Debug(azureCtx, "Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudAzure
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.Region = obj.region
			mainStateDocument.CloudInfra = &storageTypes.InfrastructureState{
				Azure: &storageTypes.StateConfigurationAzure{},
			}
			mainStateDocument.CloudInfra.Azure.B.KubernetesVer = obj.metadata.k8sVersion
			mainStateDocument.CloudInfra.Azure.B.KubernetesDistro = string(obj.metadata.k8sName)
		}

	case consts.OperationDelete:
		if errLoadState != nil {
			return log.NewError(azureCtx, "no cluster state found", "Reason", errLoadState)
		}
		log.Debug(azureCtx, "Delete resource(s)")

	case consts.OperationGet:
		if errLoadState != nil {
			return log.NewError(azureCtx, "no cluster state found", "Reason", errLoadState)
		}
		log.Debug(azureCtx, "Get storage")
	default:
		return log.NewError(azureCtx, "Invalid operation for init state")
	}

	if err := obj.client.InitClient(storage); err != nil {
		return err
	}

	// added the resource grp and region for easy of use for the client library
	obj.client.SetRegion(obj.region)
	obj.client.SetResourceGrp(obj.resourceGroup)

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	log.Debug(azureCtx, "init cloud state")

	return nil
}

func (cloud *AzureProvider) Credential(storage types.StorageFactory) error {

	log.Print(azureCtx, "Enter your SUBSCRIPTION ID")
	skey, err := helpers.UserInputCredentials(azureCtx, log)
	if err != nil {
		return err
	}

	log.Print(azureCtx, "Enter your TENANT ID")
	tid, err := helpers.UserInputCredentials(azureCtx, log)
	if err != nil {
		return err
	}

	log.Print(azureCtx, "Enter your CLIENT ID")
	cid, err := helpers.UserInputCredentials(azureCtx, log)
	if err != nil {
		return err
	}

	log.Print(azureCtx, "Enter your CLIENT SECRET")
	cs, err := helpers.UserInputCredentials(azureCtx, log)
	if err != nil {
		return err
	}

	apiStore := &storageTypes.CredentialsDocument{
		InfraProvider: consts.CloudAzure,
		Azure: &storageTypes.CredentialsAzure{
			SubscriptionID: skey,
			TenantID:       tid,
			ClientID:       cid,
			ClientSecret:   cs,
		},
	}

	// FIXME: add ping pong for validation of credentials
	//if err = os.Setenv("AZURE_SUBSCRIPTION_ID", skey); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_TENANT_ID", tid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_ID", cid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_SECRET", cs); err != nil {
	//	return err
	//}
	// ADD SOME PING method to validate credentials

	if err := storage.WriteCredentials(consts.CloudAzure, apiStore); err != nil {
		return err
	}

	return nil
}

func NewClient(
	parentCtx context.Context,
	meta types.Metadata,
	parentLogger types.LoggerFactory,
	state *storageTypes.StorageDocument,
	ClientOption func() AzureGo) (*AzureProvider, error) {

	log = parentLogger
	azureCtx = context.WithValue(parentCtx, consts.ContextModuleNameKey, string(consts.CloudAzure))

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

	log.Debug(azureCtx, "Printing", "AzureProvider", obj)

	return obj, nil
}

// Name it will contain the name of the resource to be created
func (cloud *AzureProvider) Name(resName string) types.CloudFactory {

	if err := helpers.IsValidName(azureCtx, log, resName); err != nil {
		log.Error(azureCtx, err.Error())
		return nil
	}

	cloud.chResName <- resName
	return cloud
}

// Role it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *AzureProvider) Role(resRole consts.KsctlRole) types.CloudFactory {

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		cloud.chRole <- resRole
		return cloud
	default:
		log.Error(azureCtx, "invalid role assumed", "role", string(resRole))

		return nil
	}
}

// VMType it will contain which vmType to create
func (cloud *AzureProvider) VMType(size string) types.CloudFactory {

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error(azureCtx, err.Error())
		return nil
	}
	cloud.chVMType <- size

	return cloud
}

// Visibility whether to have the resource as public or private (i.e. VMs)
func (cloud *AzureProvider) Visibility(toBePublic bool) types.CloudFactory {
	cloud.metadata.public = toBePublic
	return cloud
}

func (cloud *AzureProvider) Application(s []string) (externalApps bool) {
	return true
}

// CNI Why will be installed because it will be done by the extensions
func (cloud *AzureProvider) CNI(s string) (externalCNI bool) {

	log.Debug(azureCtx, "Printing", "cni", s)

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

// NoOfControlPlane implements types.CloudFactory.
func (obj *AzureProvider) NoOfControlPlane(no int, setter bool) (int, error) {

	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
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

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes", mainStateDocument.CloudInfra.Azure.InfoControlPlanes)
		return -1, nil
	}
	return -1, log.NewError(azureCtx, "constrains for no of controlplane >= 3 and odd number")
}

// NoOfDataStore implements types.CloudFactory.
func (obj *AzureProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoDatabase.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase.Names", mainStateDocument.CloudInfra.Azure.InfoDatabase.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoDatabase.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
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

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase", mainStateDocument.CloudInfra.Azure.InfoDatabase)
		return -1, nil
	}
	return -1, log.NewError(azureCtx, "constrains for no of Datastore>= 3 and odd number")
}

// NoOfWorkerPlane implements types.CloudFactory.
func (obj *AzureProvider) NoOfWorkerPlane(storage types.StorageFactory, no int, setter bool) (int, error) {
	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
		}
		if mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names == nil {
			// NOTE: returning nil as in case of azure the controlplane [] of instances are not initialized
			// it happens when the resource groups and network is created but interrup occurs before setter is called
			return -1, nil
		}
		log.Debug(azureCtx, "Prnting", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, log.NewError(azureCtx, "state init not called")
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

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes)

		return -1, nil
	}
	return -1, log.NewError(azureCtx, "constrains for no of workplane >= 0")
}

func GetRAWClusterInfos(storage types.StorageFactory) ([]cloudcontrolres.AllClusterData, error) {

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
			log.Debug(azureCtx, "Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func isPresent(storage types.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) error {
	err := storage.AlreadyCreated(consts.CloudAzure, region, name, ksctlClusterType)
	if err != nil {
		return log.NewError(azureCtx, "Cluster not found", "ErrStorage", err)
	}
	return nil
}

func (obj *AzureProvider) IsPresent(storage types.StorageFactory) error {

	if obj.haCluster {
		return isPresent(storage, consts.ClusterTypeHa, obj.clusterName, obj.region)
	}
	return isPresent(storage, consts.ClusterTypeMang, obj.clusterName, obj.region)
}
