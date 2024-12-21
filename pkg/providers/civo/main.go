// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package civo

import (
	"context"
	"encoding/json"
	"strings"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
	cloud_control_res "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

func (*CivoProvider) GetStateFile(types.StorageFactory) (string, error) {
	cloudstate, err := json.Marshal(mainStateDocument)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(
			log.NewError(civoCtx, "failed to serialize the state", "Reason", err),
		)
	}

	return string(cloudstate), nil
}

func (client *CivoProvider) GetStateForHACluster(storage types.StorageFactory) (cloud_control_res.CloudResourceState, error) {

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
	log.Debug(civoCtx, "Printing", "cloudState", payload)
	log.Success(civoCtx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (obj *CivoProvider) InitState(storage types.StorageFactory, operation consts.KsctlOperation) error {

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
			return ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(civoCtx, "cluster already exist", "name", mainStateDocument.ClusterName, "region", mainStateDocument.Region),
			)
		}

		if errLoadState == nil && !mainStateDocument.CloudInfra.Civo.B.IsCompleted {
			// file present but not completed
			log.Debug(civoCtx, "RESUME triggered!!")
		} else {
			log.Debug(civoCtx, "Fresh state!!")

			mainStateDocument.ClusterName = obj.clusterName
			mainStateDocument.InfraProvider = consts.CloudCivo
			mainStateDocument.Region = obj.region
			mainStateDocument.ClusterType = string(clusterType)
			mainStateDocument.CloudInfra = &storageTypes.InfrastructureState{
				Civo: &storageTypes.StateConfigurationCivo{},
			}
			mainStateDocument.CloudInfra.Civo.B.KubernetesVer = obj.k8sVersion
		}

	case consts.OperationGet:

		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(civoCtx, "Get storage")

	case consts.OperationDelete:

		if errLoadState != nil {
			return errLoadState
		}
		log.Debug(civoCtx, "Delete resource(s)")
	default:
		return ksctlErrors.ErrInvalidOperation.Wrap(
			log.NewError(civoCtx, "Invalid operation for init state"),
		)
	}

	if err := obj.client.InitClient(storage, obj.region); err != nil {
		return err
	}

	if err := validationOfArguments(obj); err != nil {
		return err
	}
	log.Debug(civoCtx, "init cloud state")
	return nil
}

func (cloud *CivoProvider) Credential(storage types.StorageFactory) error {

	log.Print(civoCtx, "Enter CIVO TOKEN")
	token, err := helpers.UserInputCredentials(civoCtx, log)
	if err != nil {
		return err
	}
	client, err := civogo.NewClient(token, "LON1")
	if err != nil {
		return err
	}
	id := client.GetAccountID()

	if len(id) == 0 {
		return ksctlErrors.ErrFailedCloudAccountAuth.Wrap(
			log.NewError(civoCtx, "Invalid user"),
		)
	}
	log.Print(civoCtx, "Recieved accountId", "userId", id)

	if err := storage.WriteCredentials(consts.CloudCivo,
		&storageTypes.CredentialsDocument{
			InfraProvider: consts.CloudCivo,
			Civo:          &storageTypes.CredentialsCivo{Token: token},
		}); err != nil {
		return err
	}

	return nil
}

func NewClient(parentCtx context.Context, meta types.Metadata, parentLogger types.LoggerFactory, state *storageTypes.StorageDocument, ClientOption func() CivoGo) (*CivoProvider, error) {
	log = parentLogger

	civoCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.CloudCivo))

	mainStateDocument = state

	obj := &CivoProvider{
		clusterName: meta.ClusterName,
		region:      meta.Region,
		haCluster:   meta.IsHA,
		metadata: metadata{
			k8sVersion: meta.K8sVersion,
		},
		client: ClientOption(),
	}
	log.Debug(civoCtx, "Printing", "CivoProvider", obj)
	return obj, nil
}

func (cloud *CivoProvider) Name(resName string) types.CloudFactory {

	if err := helpers.IsValidName(civoCtx, log, resName); err != nil {
		log.Error("Resource Name", err.Error())
		return nil
	}
	cloud.chResName <- resName
	return cloud
}

func (cloud *CivoProvider) Role(resRole consts.KsctlRole) types.CloudFactory {

	if !helpers.ValidateRole(resRole) {
		log.Error("invalidParameters",
			ksctlErrors.ErrInvalidKsctlRole.Wrap(
				log.NewError(civoCtx, "invalid role", "role", resRole)).
				Error(),
		)
		return nil
	}

	cloud.chRole <- resRole
	log.Debug(civoCtx, "Printing", "Role", resRole)
	return cloud
}

func (cloud *CivoProvider) VMType(size string) types.CloudFactory {

	if err := isValidVMSize(cloud, size); err != nil {
		log.Error("VM", err.Error())
		return nil
	}
	cloud.chVMType <- size
	log.Debug(civoCtx, "Printing", "VMSize", size)
	return cloud
}

func (cloud *CivoProvider) Visibility(toBePublic bool) types.CloudFactory {
	cloud.metadata.public = toBePublic
	log.Debug(civoCtx, "Printing", "willBePublic", toBePublic)
	return cloud
}

func (cloud *CivoProvider) SupportForApplications() bool {
	return true
}

func aggregratedApps(s []string) (ret string) {
	if len(s) == 0 {
		ret = "traefik2-nodeport,metrics-server" // default: applications
	} else {
		ret = strings.Join(s, ",") + ",traefik2-nodeport,metrics-server"
	}
	log.Debug(civoCtx, "Printing", "apps", ret)
	return
}

func (client *CivoProvider) Application(s []string) (externalApps bool) {
	client.metadata.apps = aggregratedApps(s)
	return false
}

func (client *CivoProvider) CNI(s string) (externalCNI bool) {

	log.Debug(civoCtx, "Printing", "cni", s)
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

func k8sVersion(obj *CivoProvider, ver string) (string, error) {
	if len(ver) == 0 {
		return "1.26.4-k3s1", nil
	}

	ver = ver + "-k3s1"
	if err := isValidK8sVersion(obj, ver); err != nil {
		return "", err
	}
	log.Debug(civoCtx, "Printing", "k8sVersion", ver)
	return ver, nil
}

func (obj *CivoProvider) ManagedK8sVersion(ver string) types.CloudFactory {
	v, err := k8sVersion(obj, ver)
	if err != nil {
		log.Error("Managed k8s version", err.Error())
		return nil
	}
	obj.metadata.k8sVersion = v
	return obj
}

func (*CivoProvider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
	log.Debug(civoCtx, "Printing", "hostnameOfWorkerPlanes", hostnames)
	return hostnames
}

func (obj *CivoProvider) NoOfControlPlane(no int, setter bool) (int, error) {
	log.Debug(civoCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs == nil {
			return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
				log.NewError(civoCtx, "unable to fetch controlplane instanceIDs"),
			)
		}
		log.Debug(civoCtx, "Printing", "InstanceIDsOfControlplanes", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		return len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noCP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}

		currLen := len(mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMSizes = make([]string, no)
		}
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PublicIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.PrivateIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.Hostnames)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMSizes", mainStateDocument.CloudInfra.Civo.InfoControlPlanes.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfControlplane.Wrap(
		log.NewError(civoCtx, "constrains for no of controlplane >= 3 and odd number"),
	)
}

func (obj *CivoProvider) NoOfDataStore(no int, setter bool) (int, error) {
	log.Debug(civoCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs == nil {
			return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
				log.NewError(civoCtx, "unable to fetch DataStore instanceID"),
			)
		}

		log.Debug(civoCtx, "Printing", "InstanceIDsOfDatabaseNode", mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)

		return len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		obj.metadata.noDS = no

		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}

		currLen := len(mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)
		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoDatabase.VMSizes = make([]string, no)
		}

		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs", mainStateDocument.CloudInfra.Civo.InfoDatabase.VMIDs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoDatabase.PublicIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoDatabase.PrivateIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames", mainStateDocument.CloudInfra.Civo.InfoDatabase.Hostnames)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoDatabase.VMSizes", mainStateDocument.CloudInfra.Civo.InfoDatabase.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfDatastore.Wrap(
		log.NewError(civoCtx, "constrains for no of Datastore>= 3 and odd number"),
	)
}

func (obj *CivoProvider) NoOfWorkerPlane(storage types.StorageFactory, no int, setter bool) (int, error) {
	log.Debug(civoCtx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}
		if mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs == nil {
			return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
				log.NewError(civoCtx, "unable to fetch WorkerNode instanceIDs"),
			)
		}

		log.Debug(civoCtx, "Printing", "InstanceIDsOfWorkerPlane", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		return len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs), nil
	}
	if no >= 0 {
		obj.metadata.noWP = no
		if mainStateDocument == nil {
			return -1, ksctlErrors.ErrInvalidOperation.Wrap(
				log.NewError(civoCtx, "state init not called!"),
			)
		}
		currLen := len(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		newLen := no

		if currLen == 0 {
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = make([]string, no)
			mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = make([]string, no)
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
					mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = append(mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes, "")
				}
			} else {
				// for downscaling
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[:newLen]
				mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}
		err := storage.Write(mainStateDocument)
		if err != nil {
			return -1, err
		}

		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
		log.Debug(civoCtx, "Printing", "mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes", mainStateDocument.CloudInfra.Civo.InfoWorkerPlanes.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.ErrInvalidNoOfWorkerplane.Wrap(
		log.NewError(civoCtx, "constrains for no of workerplane >= 0"),
	)
}

func (obj *CivoProvider) GetRAWClusterInfos(storage types.StorageFactory) ([]cloud_control_res.AllClusterData, error) {

	var data []cloud_control_res.AllClusterData

	clusters, err := storage.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudCivo),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	convertToAllClusterDataType := func(st *storageTypes.StorageDocument, r consts.KsctlRole) (v []cloud_control_res.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Civo.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloud_control_res.VMData{
					VMID:       o.VMIDs[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Civo.FirewallIDControlPlanes,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
				})
			}

		case consts.RoleWp:
			o := st.CloudInfra.Civo.InfoWorkerPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloud_control_res.VMData{
					VMID:       o.VMIDs[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Civo.FirewallIDWorkerNodes,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
				})
			}

		case consts.RoleDs:
			o := st.CloudInfra.Civo.InfoDatabase
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, cloud_control_res.VMData{
					VMID:       o.VMIDs[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Civo.FirewallIDDatabaseNodes,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
				})
			}

		default:
			v = append(v, cloud_control_res.VMData{
				VMID:       st.CloudInfra.Civo.InfoLoadBalancer.VMID,
				VMSize:     st.CloudInfra.Civo.InfoLoadBalancer.VMSize,
				FirewallID: st.CloudInfra.Civo.FirewallIDLoadBalancer,
				PrivateIP:  st.CloudInfra.Civo.InfoLoadBalancer.PrivateIP,
				PublicIP:   st.CloudInfra.Civo.InfoLoadBalancer.PublicIP,
			})
		}
		return v
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, cloud_control_res.AllClusterData{
				CloudProvider: consts.CloudCivo,
				ClusterType:   K,
				Name:          v.ClusterName,
				Region:        v.Region,
				CP:            convertToAllClusterDataType(v, consts.RoleCp),
				WP:            convertToAllClusterDataType(v, consts.RoleWp),
				DS:            convertToAllClusterDataType(v, consts.RoleDs),
				LB:            convertToAllClusterDataType(v, consts.RoleLb)[0],

				Mgt: cloud_control_res.VMData{
					VMSize: v.CloudInfra.Civo.ManagedNodeSize,
				},
				ManagedK8sID: v.CloudInfra.Civo.ManagedClusterID,
				NetworkID:    v.CloudInfra.Civo.NetworkID,
				SSHKeyName:   v.CloudInfra.Civo.B.SSHKeyName,
				SSHKeyID:     v.CloudInfra.Civo.B.SSHID,

				NoWP:  len(v.CloudInfra.Civo.InfoWorkerPlanes.VMIDs),
				NoCP:  len(v.CloudInfra.Civo.InfoControlPlanes.VMIDs),
				NoDS:  len(v.CloudInfra.Civo.InfoDatabase.VMIDs),
				NoMgt: v.CloudInfra.Civo.NoManagedNodes,

				K8sDistro: v.BootstrapProvider,
				HAProxyVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeHa) {
						return v.K8sBootstrap.B.HAProxyVersion
					}
					return ""
				}(),
				EtcdVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeHa) {
						return v.K8sBootstrap.B.EtcdVersion
					}
					return ""
				}(),
				K8sVersion: func() string {
					switch v.BootstrapProvider {
					case consts.K8sK3s:
						return v.K8sBootstrap.K3s.K3sVersion
					case consts.K8sKubeadm:
						return v.K8sBootstrap.Kubeadm.KubeadmVersion
					default:
						return v.CloudInfra.Civo.B.KubernetesVer
					}
				}(),
				Apps: func() (_a []string) {
					for _, a := range v.Addons.Apps {
						_a = append(_a, a.String())
					}
					return
				}(),
				Cni: v.Addons.Cni.String(),
			})
			log.Debug(civoCtx, "Printing", "cloudClusterInfoFetched", data)
		}
	}

	return data, nil
}

func isPresent(storage types.StorageFactory, ksctlClusterType consts.KsctlClusterType, name, region string) error {
	err := storage.AlreadyCreated(consts.CloudCivo, region, name, ksctlClusterType)
	if err != nil {
		return err
	}
	return nil
}

func (obj *CivoProvider) IsPresent(storage types.StorageFactory) error {
	if obj.haCluster {
		return isPresent(storage, consts.ClusterTypeHa, obj.clusterName, obj.region)
	}
	return isPresent(storage, consts.ClusterTypeMang, obj.clusterName, obj.region)
}

func (obj *CivoProvider) GetKubeconfig(storage types.StorageFactory) (*string, error) {
	_read, err := storage.Read()
	if err != nil {
		log.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig := _read.ClusterKubeConfig
	return &kubeconfig, nil
}
