package azure

import (
	"context"
	"encoding/json"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
	cloudcontrolres "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

func (*AzureProvider) GetStateFile(types.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(azureCtx, "failed to serialize the state", "Reason", err),
		)
	}
	log.Debug(azureCtx, "Printing", "cloudstate", cloudstate)
	return string(cloudstate), nil
}

func (*AzureProvider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Hostnames)
	log.Debug(azureCtx, "Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (obj *AzureProvider) ManagedK8sVersion(ver string) types.CloudFactory {
	log.Debug(azureCtx, "Printing", "K8sVersion", ver)
	if err := isValidK8sVersion(obj, ver); err != nil {
		log.Error("Managed k8s version", err.Error())
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

	obj.resourceGroup = generateResourceGroupName(obj.clusterName, string(clusterType))

	errLoadState := loadStateHelper(storage)
	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && mainStateDocument.CloudInfra.Azure.B.IsCompleted {
			return ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(azureCtx, "cluster already exist", "name", mainStateDocument.ClusterName, "region", mainStateDocument.Region),
			)
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
		}

	case consts.OperationDelete:
		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(azureCtx, "Delete resource(s)")

	case consts.OperationGet:
		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(azureCtx, "Get storage")
	default:
		return ksctlErrors.ErrInvalidOperation.Wrap(
			log.NewError(azureCtx, "Invalid operation for init state"),
		)
	}

	if err := obj.client.InitClient(storage); err != nil {
		return err
	}

	if err := validationOfArguments(obj); err != nil {
		return err
	}

	obj.client.SetRegion(obj.region)
	obj.client.SetResourceGrp(obj.resourceGroup)

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
	azureCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.CloudAzure))

	mainStateDocument = state

	obj := &AzureProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
		},
		client: ClientOption(),
	}

	log.Debug(azureCtx, "Printing", "AzureProvider", obj)

	return obj, nil
}

func (cloud *AzureProvider) Name(resName string) types.CloudFactory {

	if err := helpers.IsValidName(azureCtx, log, resName); err != nil {
		log.Error("Resource Name", err.Error())
		return nil
	}

	cloud.chResName <- resName
	return cloud
}

func (cloud *AzureProvider) Role(resRole consts.KsctlRole) types.CloudFactory {

	if !helpers.ValidateRole(resRole) {
		log.Error("invalidParameters",
			ksctlErrors.ErrInvalidKsctlRole.Wrap(
				log.NewError(azureCtx, "invalid role", "role", resRole)).
				Error(),
		)
		return nil
	}

	cloud.chRole <- resRole
	log.Debug(azureCtx, "Printing", "Role", resRole)
	return cloud
}

func (cloud *AzureProvider) VMType(size string) types.CloudFactory {

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error("VM", err.Error())
		return nil
	}
	cloud.chVMType <- size

	return cloud
}

func (cloud *AzureProvider) Visibility(toBePublic bool) types.CloudFactory {
	cloud.metadata.public = toBePublic
	return cloud
}

func (cloud *AzureProvider) Application(s []string) (externalApps bool) {
	return true
}

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

func (obj *AzureProvider) NoOfControlPlane(no int, setter bool) (int, error) {

	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names == nil {
			return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
				log.NewError(azureCtx, "unable to fetch controlplane instanceIDs"),
			)
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
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
			mainStateDocument.CloudInfra.Azure.InfoControlPlanes.VMSizes = make([]string, no)
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoControlPlanes", mainStateDocument.CloudInfra.Azure.InfoControlPlanes)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
		log.NewError(azureCtx, "constrains for no of controlplane >= 3 and odd number"),
	)
}

func (obj *AzureProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Azure.InfoDatabase.Names == nil {
			return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
				log.NewError(azureCtx, "unable to fetch DataStore instanceID"),
			)
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase.Names", mainStateDocument.CloudInfra.Azure.InfoDatabase.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoDatabase.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
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
			mainStateDocument.CloudInfra.Azure.InfoDatabase.VMSizes = make([]string, no)
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoDatabase", mainStateDocument.CloudInfra.Azure.InfoDatabase)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
		log.NewError(azureCtx, "constrains for no of Datastore>= 3 and odd number"),
	)
}

func (obj *AzureProvider) NoOfWorkerPlane(storage types.StorageFactory, no int, setter bool) (int, error) {
	log.Debug(azureCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names == nil {
			return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
				log.NewError(azureCtx, "unable to fetch WorkerNode instanceIDs"),
			)
		}
		log.Debug(azureCtx, "Prnting", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names)
		return len(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(azureCtx, "state init not called!"),
			)
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
			mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = make([]string, no)
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
					mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = append(mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes, "")
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
				mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}

		if err := storage.Write(mainStateDocument); err != nil {
			return -1, err
		}

		log.Debug(azureCtx, "Printing", "mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes", mainStateDocument.CloudInfra.Azure.InfoWorkerPlanes)

		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
		log.NewError(azureCtx, "constrains for no of workerplane >= 0"),
	)
}

func (obj *AzureProvider) GetRAWClusterInfos(storage types.StorageFactory) ([]cloudcontrolres.AllClusterData, error) {

	var data []cloudcontrolres.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAzure),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	convertToAllClusterDataType := func(st *storageTypes.StorageDocument, r consts.KsctlRole) (v []cloudcontrolres.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Azure.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMName:       o.Names[i],
					VMSize:       o.VMSizes[i],
					FirewallID:   st.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupID,
					FirewallName: st.CloudInfra.Azure.InfoControlPlanes.NetworkSecurityGroupName,
					PublicIP:     o.PublicIPs[i],
					PrivateIP:    o.PrivateIPs[i],
					SubnetID:     st.CloudInfra.Azure.SubnetID,
					SubnetName:   st.CloudInfra.Azure.SubnetName,
				})
			}

		case consts.RoleWp:
			o := st.CloudInfra.Azure.InfoWorkerPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMName:       o.Names[i],
					VMSize:       o.VMSizes[i],
					FirewallID:   st.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupID,
					FirewallName: st.CloudInfra.Azure.InfoWorkerPlanes.NetworkSecurityGroupName,
					PublicIP:     o.PublicIPs[i],
					PrivateIP:    o.PrivateIPs[i],
					SubnetID:     st.CloudInfra.Azure.SubnetID,
					SubnetName:   st.CloudInfra.Azure.SubnetName,
				})
			}

		case consts.RoleDs:
			o := st.CloudInfra.Azure.InfoDatabase
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloudcontrolres.VMData{
					VMName:       o.Names[i],
					VMSize:       o.VMSizes[i],
					FirewallID:   st.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupID,
					FirewallName: st.CloudInfra.Azure.InfoDatabase.NetworkSecurityGroupName,
					PublicIP:     o.PublicIPs[i],
					PrivateIP:    o.PrivateIPs[i],
					SubnetID:     st.CloudInfra.Azure.SubnetID,
					SubnetName:   st.CloudInfra.Azure.SubnetName,
				})
			}

		default:
			v = append(v, cloudcontrolres.VMData{
				VMName:       st.CloudInfra.Azure.InfoLoadBalancer.Name,
				VMSize:       st.CloudInfra.Azure.InfoLoadBalancer.VMSize,
				FirewallID:   st.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupID,
				FirewallName: st.CloudInfra.Azure.InfoLoadBalancer.NetworkSecurityGroupName,
				PublicIP:     st.CloudInfra.Azure.InfoLoadBalancer.PublicIP,
				PrivateIP:    st.CloudInfra.Azure.InfoLoadBalancer.PrivateIP,
				SubnetID:     st.CloudInfra.Azure.SubnetID,
				SubnetName:   st.CloudInfra.Azure.SubnetName,
			})
		}
		return v
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloudcontrolres.AllClusterData{
				CloudProvider: consts.CloudAzure,
				Name:          v.ClusterName,
				Region:        v.Region,
				ClusterType:   K,
				CP:            convertToAllClusterDataType(v, consts.RoleCp),
				WP:            convertToAllClusterDataType(v, consts.RoleWp),
				DS:            convertToAllClusterDataType(v, consts.RoleDs),
				LB:            convertToAllClusterDataType(v, consts.RoleLb)[0],

				NoWP:  len(v.CloudInfra.Azure.InfoWorkerPlanes.Names),
				NoCP:  len(v.CloudInfra.Azure.InfoControlPlanes.Names),
				NoDS:  len(v.CloudInfra.Azure.InfoDatabase.Names),
				NoMgt: v.CloudInfra.Azure.NoManagedNodes,
				Mgt: cloudcontrolres.VMData{
					VMSize: v.CloudInfra.Azure.ManagedNodeSize,
				},
				ManagedK8sName:  v.CloudInfra.Azure.ManagedClusterName,
				NetworkName:     v.CloudInfra.Azure.VirtualNetworkName,
				NetworkID:       v.CloudInfra.Azure.VirtualNetworkID,
				ResourceGrpName: v.CloudInfra.Azure.ResourceGroupName,
				SSHKeyName:      v.CloudInfra.Azure.B.SSHKeyName,
				SSHKeyID:        v.CloudInfra.Azure.B.SSHID,

				K8sDistro:  v.BootstrapProvider,
				K8sVersion: v.CloudInfra.Azure.B.KubernetesVer,
				Apps: func() (_a []string) {
					for _, a := range v.Addons.Apps {
						_a = append(_a, a.String())
					}
					return
				}(),
				Cni: v.Addons.Cni.String(),
			})
			log.Debug(azureCtx, "Printing", "cloudClusterInfoFetched", data)

		}
	}

	return data, nil
}

func isPresent(storage types.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) error {
	err := storage.AlreadyCreated(consts.CloudAzure, region, name, ksctlClusterType)
	if err != nil {
		return err
	}
	return nil
}

func (obj *AzureProvider) IsPresent(storage types.StorageFactory) error {

	if obj.haCluster {
		return isPresent(storage, consts.ClusterTypeHa, obj.clusterName, obj.region)
	}
	return isPresent(storage, consts.ClusterTypeMang, obj.clusterName, obj.region)
}

func (obj *AzureProvider) GetKubeconfig(storage types.StorageFactory) (*string, error) {
	_read, err := storage.Read()
	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}
	log.Debug(azureCtx, "data", "read", _read)

	kubeconfig := _read.ClusterKubeConfig
	log.Debug(azureCtx, "data", "kubeconfig", kubeconfig)
	return &kubeconfig, nil
}
