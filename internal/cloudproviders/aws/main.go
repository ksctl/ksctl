package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"

	cloudcontrolres "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	mainStateDocument *types.StorageDocument
	clusterType       consts.KsctlClusterType
	ctx               context.Context
	log               resources.LoggerFactory
	clusterDirName    string
)

type Credential struct {
	AccessKeyID string `json:"access_key_id"`
	Secret      string `json:"secret_access_key"`
}

type AwsProvider struct {
	clusterName string `json:"cluster_name"`
	haCluster   bool   `json:"ha_cluster"`
	region      string `json:"region"`
	vpc         string `json:"vpc"`
	accessKeyID string `json:"access_key_id"`
	secret      string `json:"secret_access_key"`
	session     aws.Config
	metadata    metadata

	mu sync.Mutex

	nicprint  sync.Once
	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client  AwsGo
	SSHPath string `json:"ssh_key"`
}

func (*AwsProvider) IsPresent(resources.StorageFactory) error {
	return nil
}

func (*AwsProvider) Version(string) resources.CloudFactory {
	panic("unimplemented")
}

type metadata struct {
	public  bool
	cni     string
	version string

	noCP int
	noWP int
	noDS int

	k8sName    consts.KsctlKubernetes
	k8sVersion string
}

const (
	FILE_PERM_CLUSTER_DIR        = os.FileMode(0750)
	FILE_PERM_CLUSTER_STATE      = os.FileMode(0640)
	FILE_PERM_CLUSTER_KUBECONFIG = os.FileMode(0755)
	STATE_FILE_NAME              = string("cloud-state.json")
	KUBECONFIG_FILE_NAME         = string("kubeconfig")
)

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

		session: NEWCLIENT(),
		client:  ClientOption(),
	}

	log.Debug("Printing", "AwsProvider", obj)

	return obj, nil
}

func NEWCLIENT() aws.Config {
	NewSession, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")),
	)
	if err != nil {
		log.Error(err.Error())
	}
	log.Success("AWS Session created successfully")

	return NewSession
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
	clusterDirName = obj.clusterName + " " + obj.vpc + " " + obj.region
	log.Print("clusterDirName", "", clusterDirName)

	errLoadState := loadStateHelper(storage)

	switch opration {
	case consts.OperationStateCreate:
		if errLoadState == nil && mainStateDocument.CloudInfra.Aws.IsCompleted {
			return fmt.Errorf("cluster %s already exists", obj.clusterName)
		}
		if errLoadState == nil && !mainStateDocument.CloudInfra.Aws.IsCompleted {
			log.Warn("Cluster state found but not completed, resuming operation")
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

	case consts.OperationStateDelete:
		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Delete resource(s)")

	case consts.OperationStateGet:
		if errLoadState != nil {
			return log.NewError("no cluster state found reason:%s\n", errLoadState.Error())
		}
		log.Debug("Get resources")
	default:
		return log.NewError("Invalid operation for init state")
	}

	if err := obj.client.InitClient(storage); err != nil {
		fmt.Errorf("failed to initialize aws client reason:%s\n", err.Error())
	}

	obj.client.SetRegion(obj.region)
	obj.client.SetVpc(obj.vpc)

	if err := validationOfArguments(obj); err != nil {
		return log.NewError(err.Error())
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
		IPv4ControlPlanes: mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs,
		IPv4DataStores:    mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs,
		IPv4WorkerPlanes:  mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs,
		IPv4LoadBalancer:  mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP,

		PrivateIPv4ControlPlanes: mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs,
		PrivateIPv4DataStores:    mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs,
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

//func (obj *AwsProvider) SupportForCNI() bool {
//	//TODO implement me
//	fmt.Println("AWS Support for CNI")
//	return false
//
//}

func (obj *AwsProvider) Application(s string) bool {
	return true
}

func (obj *AwsProvider) CNI(s string) bool {
	obj.metadata.cni = string(consts.CNINone)
	return true
}

func (obj *AwsProvider) NoOfWorkerPlane(storage resources.StorageFactory, no int, setter bool) (int, error) {
	log.Print("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names == nil {
			return -1, nil
		}
		return len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		currLen := len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names = make([]string, no)
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
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.DiskNames, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs, "")
				}
			} else {
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names[:newLen]
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
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names == nil {
			return -1, nil
		}

		log.Debug("Printing", "mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names)
		return len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names = make([]string, no)
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
	log.Print("Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.Names == nil {
			return -1, nil
		}

		log.Print("Printing", "awsCloudState.InfoDatabase.Names", mainStateDocument.CloudInfra.Aws.InfoDatabase.Names)
		return len(mainStateDocument.CloudInfra.Aws.InfoDatabase.Names), nil
	}
	if no >= 1 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, log.NewError("state init not called")
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoDatabase.Names)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoDatabase.Names = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.DiskNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
		}

		log.Debug("Printing", "awsCloudState.InfoDatabase", mainStateDocument.CloudInfra.Aws.InfoDatabase)
		return -1, nil
	}
	return -1, log.NewError("constrains for no of Datastore>= 1 and odd number")
}

func (obj *AwsProvider) GetHostNameAllWorkerNode() []string {
	var hostnames []string = make([]string, len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names))
	copy(hostnames, mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names)
	log.Debug("Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (obj *AwsProvider) SwitchCluster(factory resources.StorageFactory) error {
	//TODO implement me
	fmt.Println("AWS Switch Cluster")
	return nil

}

func (obj *AwsProvider) GetStateFile(factory resources.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", err
	}
	log.Debug("Printing", "cloudstate", cloudstate)
	return string(cloudstate), nil

}

func (obj *AwsProvider) GetSecretTokens(factory resources.StorageFactory) (map[string][]byte, error) {

	acesskeyid := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	return map[string][]byte{
		"aws_access_key_id":     []byte(acesskeyid),
		"aws_secret_access_key": []byte(secret),
	}, nil
}

func GetRAWClusterInfos(storage resources.StorageFactory, meta resources.Metadata) ([]cloudcontrolres.AllClusterData, error) {

	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

	var data []cloudcontrolres.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[string]string{
		"cloud":       string(consts.CloudAws),
		"clusterType": "",
	})
	if err != nil {
		return nil, err
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudcontrolres.AllClusterData{
				Provider: consts.CloudAws,
				Name:     v.ClusterName,
				Region:   v.Region,
				Type:     K,

				NoWP:  len(v.CloudInfra.Aws.InfoWorkerPlanes.Names),
				NoCP:  len(v.CloudInfra.Aws.InfoControlPlanes.Names),
				NoDS:  len(v.CloudInfra.Aws.InfoDatabase.Names),
				NoMgt: v.CloudInfra.Aws.NoManagedNodes,

				K8sDistro:  consts.KsctlKubernetes(v.CloudInfra.Aws.B.KubernetesDistro),
				K8sVersion: v.CloudInfra.Aws.B.KubernetesVer,
			})
			log.Debug("Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}
