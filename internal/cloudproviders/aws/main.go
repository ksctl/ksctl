package aws

import (
	"context"
	"encoding/json"
	"fmt"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"

	cloudcontrolres "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

func isPresent(storage types.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) error {

	err := storage.AlreadyCreated(consts.CloudAws, region, name, ksctlClusterType)
	if err != nil {
		return log.NewError(awsCtx, "Cluster not found", "ErrStorage", err)
	}
	return nil
}

func (obj *AwsProvider) IsPresent(storage types.StorageFactory) error {

	if obj.haCluster {
		return isPresent(storage, consts.ClusterTypeHa, obj.clusterName, obj.region)
	}
	return isPresent(storage, consts.ClusterTypeMang, obj.clusterName, obj.region)
}

func (obj *AwsProvider) ManagedK8sVersion(ver string) types.CloudFactory {
	log.Debug(awsCtx, "Printing", "K8sVersion", ver)
	if err := isValidK8sVersion(obj, ver); err != nil {
		log.Error("Managed k8s version", err.Error())
		return nil
	}

	obj.metadata.k8sVersion = ver

	return obj
}

func (cloud *AwsProvider) Credential(storage types.StorageFactory) error {
	log.Print(awsCtx, "Enter your AWS ACCESS KEY")
	acesskey, err := helpers.UserInputCredentials(awsCtx, log)
	if err != nil {
		return err
	}

	log.Print(awsCtx, "Enter your AWS SECRET KEY")
	acesskeysecret, err := helpers.UserInputCredentials(awsCtx, log)
	if err != nil {
		return err
	}

	apiStore := &storageTypes.CredentialsDocument{
		InfraProvider: consts.CloudAws,
		Aws: &storageTypes.CredentialsAws{
			AccessKeyId:     acesskey,
			SecretAccessKey: acesskeysecret,
		},
	}

	if err := storage.WriteCredentials(consts.CloudAws, apiStore); err != nil {
		return err
	}

	return nil
}

func NewClient(parentCtx context.Context,
	meta types.Metadata,
	parentLogger types.LoggerFactory,
	state *storageTypes.StorageDocument,
	ClientOption func() AwsGo) (*AwsProvider, error) {
	log = parentLogger // intentional shallow copy so that we can use the same
	// logger to be used multiple places
	awsCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.CloudAws))

	mainStateDocument = state

	obj := &AwsProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
			// k8sName:    meta.K8sDistro,
		},

		client: ClientOption(),
	}

	log.Debug(awsCtx, "Printing", "AwsProvider", obj)

	return obj, nil
}

func (obj *AwsProvider) Name(resName string) types.CloudFactory {

	if err := helpers.IsValidName(awsCtx, log, resName); err != nil {
		log.Error(err.Error())
		return nil
	}
	obj.chResName <- resName
	return obj
}

func (obj *AwsProvider) InitState(storage types.StorageFactory, opration consts.KsctlOperation) error {

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
			return ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(awsCtx, "cluster already exist", "name", mainStateDocument.ClusterName, "region", mainStateDocument.Region),
			)
		}
		if errLoadState == nil && !mainStateDocument.CloudInfra.Aws.IsCompleted {
			log.Note(awsCtx, "Cluster state found but not completed, resuming operation")
		} else {
			log.Debug(awsCtx, "Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudAws
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.Region = obj.region
			mainStateDocument.CloudInfra = &storageTypes.InfrastructureState{
				Aws: &storageTypes.StateConfigurationAws{},
			}
			mainStateDocument.CloudInfra.Aws.B.KubernetesVer = obj.metadata.k8sVersion
		}

	case consts.OperationDelete:
		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(awsCtx, "Delete resource(s)")

	case consts.OperationGet:
		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(awsCtx, "Get storage")
	default:
		return ksctlErrors.ErrInvalidOperation.Wrap(
			log.NewError(awsCtx, "Invalid operation for init state"),
		)
	}

	obj.client.SetRegion(obj.region)

	if err := obj.client.InitClient(storage); err != nil {
		return err
	}

	if err := validationOfArguments(obj); err != nil {
		return err
	}
	log.Debug(awsCtx, "init cloud state")

	return nil
}

func (obj *AwsProvider) GetStateForHACluster(storage types.StorageFactory) (cloudcontrolres.CloudResourceState, error) {

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

	log.Debug(awsCtx, "Printing", "awsStateTransferPayload", payload)

	log.Success(awsCtx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *AwsProvider) Role(resRole consts.KsctlRole) types.CloudFactory {

	if !helpers.ValidateRole(resRole) {
		log.Error("invalidParameters",
			ksctlErrors.ErrInvalidKsctlRole.Wrap(
				log.NewError(awsCtx, "invalid role", "role", resRole)).
				Error(),
		)
		return nil
	}

	obj.chRole <- resRole
	log.Debug(awsCtx, "Printing", "Role", resRole)
	return obj
}

func (obj *AwsProvider) VMType(size string) types.CloudFactory {

	if err := isValidVMSize(obj, size); err != nil {
		log.Error("VM", err.Error())
		return nil
	}
	obj.chVMType <- size

	return obj
}

func (obj *AwsProvider) Visibility(toBePublic bool) types.CloudFactory {
	obj.metadata.public = toBePublic
	return obj
}

func (obj *AwsProvider) SupportForApplications() bool {
	return false

}

func (obj *AwsProvider) Application(s []string) bool {
	return true
}

func (obj *AwsProvider) CNI(s string) (externalCNI bool) {

	log.Debug(awsCtx, "Printing", "cni", s)
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

func (obj *AwsProvider) NoOfWorkerPlane(storage types.StorageFactory, no int, setter bool) (int, error) {
	log.Debug(awsCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames == nil {
			return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
				log.NewError(awsCtx, "unable to fetch WorkerNode instanceIDs"),
			)
		}
		return len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}
		currLen := len(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = make([]string, no)
		} else {
			if currLen == newLen {
				return -1, nil
			} else if currLen < newLen {
				for i := currLen; i < newLen; i++ {
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = append(mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes, "")
				}
			} else {
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return -1, err
		}

		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
		log.NewError(awsCtx, "constrains for no of workerplane >= 0"),
	)

}

func (obj *AwsProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	if !setter {
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames == nil {
			return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
				log.NewError(awsCtx, "unable to fetch Controlplane instanceIDs"),
			)
		}

		log.Debug(awsCtx, "Printing", "mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames)
		return len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.VMSizes = make([]string, no)
		}

		log.Debug(awsCtx, "Printing", "awsCloudState.InfoControlplanes", mainStateDocument.CloudInfra.Aws.InfoControlPlanes)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
		log.NewError(awsCtx, "constrains for no of controlplane >= 3 and odd number"),
	)

}

func (obj *AwsProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug(awsCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames == nil {
			return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
				log.NewError(awsCtx, "unable to fetch Datastore instanceIDs"),
			)
		}

		log.Debug(awsCtx, "Printing", "awsCloudState.InfoDatabase.Names", mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames)
		return len(mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(awsCtx, "state init not called!"),
			)
		}

		currLen := len(mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			mainStateDocument.CloudInfra.Aws.InfoDatabase.VMSizes = make([]string, no)
		}

		log.Debug(awsCtx, "Printing", "awsCloudState.InfoDatabase", mainStateDocument.CloudInfra.Aws.InfoDatabase)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
		log.NewError(awsCtx, "constrains for no of Datastore>= 3 and odd number"),
	)
}

func (obj *AwsProvider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames)
	log.Debug(awsCtx, "Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (obj *AwsProvider) GetStateFile(factory types.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(awsCtx, "failed to serialize the state", "Reason", err),
		)
	}

	return string(cloudstate), nil
}

func (obj *AwsProvider) GetRAWClusterInfos(storage types.StorageFactory) ([]cloudcontrolres.AllClusterData, error) {

	var data []cloudcontrolres.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAws),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	convertToAllClusterDataType := func(st *storageTypes.StorageDocument, r consts.KsctlRole) (v []cloudcontrolres.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Aws.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMID:       o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   st.CloudInfra.Aws.SubnetIDs[0],
					SubnetName: st.CloudInfra.Aws.SubnetNames[0],
				})
			}

		case consts.RoleWp:
			o := st.CloudInfra.Aws.InfoWorkerPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMID:       o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   st.CloudInfra.Aws.SubnetIDs[0],
					SubnetName: st.CloudInfra.Aws.SubnetNames[0],
				})
			}

		case consts.RoleDs:
			o := st.CloudInfra.Aws.InfoDatabase
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMID:       o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   st.CloudInfra.Aws.SubnetIDs[0],
					SubnetName: st.CloudInfra.Aws.SubnetNames[0],
				})
			}

		default:
			v = append(v, cloudcontrolres.VMData{
				VMID:       st.CloudInfra.Aws.InfoLoadBalancer.InstanceID,
				VMSize:     st.CloudInfra.Aws.InfoLoadBalancer.VMSize,
				FirewallID: st.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID,
				PublicIP:   st.CloudInfra.Aws.InfoLoadBalancer.PublicIP,
				PrivateIP:  st.CloudInfra.Aws.InfoLoadBalancer.PrivateIP,
				SubnetID:   st.CloudInfra.Aws.SubnetIDs[0],
				SubnetName: st.CloudInfra.Aws.SubnetNames[0],
			})
		}
		return v
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudcontrolres.AllClusterData{
				CloudProvider: consts.CloudAws,
				Name:          v.ClusterName,
				Region:        v.Region,
				ClusterType:   K,
				CP:            convertToAllClusterDataType(v, consts.RoleCp),
				WP:            convertToAllClusterDataType(v, consts.RoleWp),
				DS:            convertToAllClusterDataType(v, consts.RoleDs),
				LB:            convertToAllClusterDataType(v, consts.RoleLb)[0],

				NoWP:  len(v.CloudInfra.Aws.InfoWorkerPlanes.HostNames),
				NoCP:  len(v.CloudInfra.Aws.InfoControlPlanes.HostNames),
				NoDS:  len(v.CloudInfra.Aws.InfoDatabase.HostNames),
				NoMgt: v.CloudInfra.Aws.NoManagedNodes,
				Mgt: cloudcontrolres.VMData{
					VMSize: v.CloudInfra.Aws.ManagedNodeSize,
				},
				ManagedK8sID: v.CloudInfra.Aws.ManagedClusterArn,
				NetworkID:    v.CloudInfra.Aws.VpcId,
				NetworkName:  v.CloudInfra.Aws.VpcName,
				SSHKeyID:     v.CloudInfra.Aws.B.SSHID,
				SSHKeyName:   v.CloudInfra.Aws.B.SSHKeyName,

				K8sDistro:  v.BootstrapProvider,
				K8sVersion: v.CloudInfra.Aws.B.KubernetesVer,
				Apps: func() (_a []string) {
					for _, a := range v.Addons.Apps {
						_a = append(_a, a.String())
					}
					return
				}(),
				Cni: v.Addons.Cni.String(),
			})
			log.Debug(awsCtx, "Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func (obj *AwsProvider) GetKubeconfig(storage types.StorageFactory) (*string, error) {

	if !obj.haCluster {
		kubeconfig, err := obj.client.GetKubeConfig(awsCtx, mainStateDocument.CloudInfra.Aws.ManagedClusterName)
		if err != nil {
			return nil, err
		}
		return &kubeconfig, nil
	}

	kubeconfig := mainStateDocument.ClusterKubeConfig
	return &kubeconfig, nil
}
