package aws

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/resources"

	cloudcontrolres "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	mainStateDocument *types.StorageDocument
	clusterType       consts.KsctlClusterType
	log               resources.LoggerFactory
)

type metadata struct {
	public bool
	cni    string
	// version string

	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
	k8sVersion string
}

type AwsProvider struct {
	clusterName string
	haCluster   bool
	region      string
	vpc         string
	metadata    metadata

	mu sync.Mutex

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client AwsGo
}

func isPresent(storage resources.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) bool {
	err := storage.AlreadyCreated(consts.CloudAws, region, name, ksctlClusterType)
	return err == nil
}

func (obj *AwsProvider) IsPresent(storage resources.StorageFactory) error {

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

func (obj *AwsProvider) Version(ver string) resources.CloudFactory {
	// TODO for ManagedCluster EKS
	return obj
}

func ReturnAwsStruct(meta resources.Metadata, state *types.StorageDocument, ClientOption func() AwsGo) (*AwsProvider, error) {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

	mainStateDocument = state

	obj := &AwsProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
			k8sName:    meta.K8sDistro,
		},

		client: ClientOption(),
	}

	obj.client.SetRegion(obj.region)
	log.Debug("Printing", "AwsProvider", obj)

	return obj, nil
}

func (obj *AwsProvider) Name(resName string) resources.CloudFactory {

	if err := helpers.IsValidName(resName); err != nil {
		log.Error(err.Error())
		return nil
	}
	obj.chResName <- resName
	return obj
}

func (obj *AwsProvider) InitState(storage resources.StorageFactory, opration consts.KsctlOperation) error {

	switch obj.haCluster {
	case false:
		clusterType = consts.ClusterTypeMang
	case true:
		clusterType = consts.ClusterTypeHa
	}

	obj.chResName = make(chan string, 1)
	obj.chRole = make(chan consts.KsctlRole, 1)
	obj.chVMType = make(chan string, 1)

	obj.vpc = fmt.Sprintf("%s-ksctl-%s-vpc", obj.clusterName, clusterType)

	errLoadState := loadStateHelper(storage)

	switch opration {
	case consts.OperationCreate:
		if errLoadState == nil && mainStateDocument.CloudInfra.Aws.IsCompleted {
			return log.NewError("cluster %s already exists", obj.clusterName)
		}
		if errLoadState == nil && !mainStateDocument.CloudInfra.Aws.IsCompleted {
			log.Note("Cluster state found but not completed, resuming operation")
		} else {
			log.Debug("Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudAws
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.Region = obj.region
			mainStateDocument.CloudInfra = &types.InfrastructureState{
				Aws: &types.StateConfigurationAws{},
			}
			mainStateDocument.CloudInfra.Aws.B.KubernetesVer = obj.metadata.k8sVersion
			mainStateDocument.CloudInfra.Aws.B.KubernetesDistro = string(obj.metadata.k8sName)
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

	if err := obj.client.InitClient(storage); err != nil {
		return err
	}

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	log.Debug("init cloud state")

	return nil
}

func (obj *AwsProvider) GetStateForHACluster(storage resources.StorageFactory) (cloudcontrolres.CloudResourceState, error) {

	payload := cloudcontrolres.CloudResourceState{
		SSHState: cloudcontrolres.SSHInfo{
			PrivateKey: mainStateDocument.SSHKeyPair.PrivateKey,
			UserName:   mainStateDocument.CloudInfra.Aws.B.SSHUser,
		},
		Metadata: cloudcontrolres.Metadata{
			ClusterName: mainStateDocument.ClusterName,
			Provider:    mainStateDocument.InfraProvider,
			Region:      mainStateDocument.Region,
			ClusterType: clusterType,
		},
		IPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP,

		PrivateIPv4ControlPlanes: utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PrivateIP,
	}

	log.Debug("Printing", "awsStateTransferPayload", payload)

	log.Success("Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *AwsProvider) NewManagedCluster(factory resources.StorageFactory, i int) error {
	//TODO implement me
	fmt.Println("AWS New Managed Cluster")
	return nil

}

func (obj *AwsProvider) DelManagedCluster(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Del Managed Cluster")
	return nil

}

func (obj *AwsProvider) Role(resRole consts.KsctlRole) resources.CloudFactory {

	switch resRole {
	case consts.RoleCp, consts.RoleDs, consts.RoleLb, consts.RoleWp:
		obj.chRole <- resRole
		return obj
	default:
		log.Error("invalid role assumed")

		return nil
	}
}

func (obj *AwsProvider) VMType(size string) resources.CloudFactory {

	if err := isValidVMSize(obj, size); err != nil {
		log.Error(err.Error())
		return nil
	}
	obj.chVMType <- size

	return obj
}

func (obj *AwsProvider) Visibility(toBePublic bool) resources.CloudFactory {
	obj.metadata.public = toBePublic
	return obj
}

func (obj *AwsProvider) SupportForApplications() bool {
	//TODO implement me
	fmt.Println("AWS Support for Applications")
	return false

}

func (obj *AwsProvider) Application(s []string) bool {
	return true
}

func (obj *AwsProvider) CNI(s string) (externalCNI bool) {

	log.Debug("Printing", "cni", s)
	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNICilium, consts.CNIFlannel:
		obj.metadata.cni = s
		return false
	case "":
		obj.metadata.cni = string(consts.CNIFlannel)
		return false
	default:
		obj.metadata.cni = string(consts.CNINone)
		return true
	}
}

func (obj *AwsProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames == nil {
			return -1, nil
		}
		return len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		currLen := len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
		} else {
			if currLen == newLen {
				return -1, nil
			} else if currLen < newLen {
				for i := currLen; i < newLen; i++ {
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs, "")
				}
			} else {
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
			}
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return -1, err
		}

		return -1, nil
	}
	return -1, log.NewError("constrains for no of workplane >= 0")

}

func (obj *AwsProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	if !setter {
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames == nil {
			return -1, nil
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames)
		return len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
		}

		log.Debug("Printing", "awsCloudState.InfoControlplanes", mainStateDocument.CloudInfra.Aws.InfoControlPlanes)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of controlplane >= 3 and odd number")

}

func (obj *AwsProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames == nil {
			return -1, nil
		}

		log.Debug("Printing", "awsCloudState.InfoDatabase.Names", mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames)
		return len(mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
		}

		log.Debug("Printing", "awsCloudState.InfoDatabase", mainStateDocument.CloudInfra.Aws.InfoDatabase)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 3 and odd number")
}

func (obj *AwsProvider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames)
	log.Debug("Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (obj *AwsProvider) GetStateFile(factory resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", log.NewError("Error marshalling state", "error", err)
	}
	log.Debug("Printing", "cloudstate", cloudstate)
	return string(cloudstate), nil

}

func GetRAWClusterInfos(storage resources.StorageFactory, meta resources.Metadata) ([]cloudcontrolres.AllClusterData, error) {

	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

	var data []cloudcontrolres.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAws),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, log.NewError("Error fetching cluster info", "error", err)
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudcontrolres.AllClusterData{
				Provider: consts.CloudAws,
				Name:     v.ClusterName,
				Region:   v.Region,
				Type:     K,

				NoWP:  len(v.CloudInfra.Aws.InfoWorkerPlanes.HostNames),
				NoCP:  len(v.CloudInfra.Aws.InfoControlPlanes.HostNames),
				NoDS:  len(v.CloudInfra.Aws.InfoDatabase.HostNames),
				NoMgt: v.CloudInfra.Aws.NoManagedNodes,

				K8sDistro:  consts.KsctlKubernetes(v.CloudInfra.Aws.B.KubernetesDistro),
				K8sVersion: v.CloudInfra.Aws.B.KubernetesVer,
			})
			log.Debug("Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}
