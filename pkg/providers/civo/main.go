// Copyright 2024 Ksctl Authors
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
	"sync"

	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/providers"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	"github.com/ksctl/ksctl/pkg/validation"

	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/utilities"
)

type Provider struct {
	l           logger.Logger
	ctx         context.Context
	state       *statefile.StorageDocument
	store       storage.Storage
	clusterType consts.KsctlClusterType
	mu          sync.Mutex

	controller.Metadata

	public bool

	// purpose: application in managed cluster
	apps string
	cni  string

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client CloudSDK
}

func NewClient(
	ctx context.Context,
	l logger.Logger,
	meta controller.Metadata,
	state *statefile.StorageDocument,
	storage storage.Storage,
	ClientOption func() CloudSDK,
) (*Provider, error) {
	p := new(Provider)
	p.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, string(consts.CloudCivo))
	p.state = state
	p.Metadata = meta
	p.l = l
	p.client = ClientOption()
	p.store = storage

	p.l.Debug(p.ctx, "Printing", "CivoProvider", p)
	return p, nil
}

func (p *Provider) GetStateFile() (string, error) {
	cloudstate, err := json.Marshal(p.state)
	if err != nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			p.l.NewError(p.ctx, "failed to serialize the state", "Reason", err),
		)
	}

	return string(cloudstate), nil
}

func (p *Provider) GetStateForHACluster() (providers.CloudResourceState, error) {

	payload := providers.CloudResourceState{
		SSHPrivateKey: p.state.SSHKeyPair.PrivateKey,
		SSHUserName:   p.state.CloudInfra.Civo.B.SSHUser,
		ClusterName:   p.ClusterName,
		Provider:      p.Provider,
		Region:        p.Region,
		ClusterType:   p.clusterType,

		// public IPs
		IPv4ControlPlanes: utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  p.state.CloudInfra.Civo.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  p.state.CloudInfra.Civo.InfoLoadBalancer.PrivateIP,
	}
	p.l.Debug(p.ctx, "Printing", "cloudState", payload)
	p.l.Success(p.ctx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (p *Provider) InitState(operation consts.KsctlOperation) error {

	if p.IsHA {
		p.clusterType = consts.ClusterTypeHa
	} else {
		p.clusterType = consts.ClusterTypeMang
	}

	p.chResName = make(chan string, 1)
	p.chRole = make(chan consts.KsctlRole, 1)
	p.chVMType = make(chan string, 1)

	errLoadState := p.loadStateHelper()

	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && p.state.CloudInfra.Civo.B.IsCompleted {
			// then found and it and the process is done then no point of duplicate creation
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				p.l.NewError(p.ctx, "cluster already exist", "name", p.state.ClusterName, "region", p.state.Region),
			)
		}

		if errLoadState == nil && !p.state.CloudInfra.Civo.B.IsCompleted {
			// file present but not completed
			p.l.Debug(p.ctx, "RESUME triggered!!")
		} else {
			p.l.Debug(p.ctx, "Fresh state!!")

			p.state.ClusterName = p.ClusterName
			p.state.InfraProvider = consts.CloudCivo
			p.state.Region = p.Region
			p.state.ClusterType = string(p.clusterType)
			p.state.CloudInfra = &statefile.InfrastructureState{
				Civo: &statefile.StateConfigurationCivo{},
			}
			p.state.CloudInfra.Civo.B.KubernetesVer = p.K8sVersion
		}

	case consts.OperationGet:

		if errLoadState != nil {
			return errLoadState
		}
		p.l.Debug(p.ctx, "Get storage")

	case consts.OperationDelete:

		if errLoadState != nil {
			return errLoadState
		}
		p.l.Debug(p.ctx, "Delete resource(s)")
	default:
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidOperation,
			p.l.NewError(p.ctx, "Invalid operation for init state"),
		)
	}

	if err := p.client.InitClient(p, p.Region); err != nil {
		return err
	}

	if err := validationOfArguments(p); err != nil {
		return err
	}
	p.l.Debug(p.ctx, "init cloud state")
	return nil
}

func (p *Provider) Credential() error {

	p.l.Print(p.ctx, "Enter CIVO TOKEN")
	token, err := validation.UserInputCredentials(p.ctx, p.l)
	if err != nil {
		return err
	}
	client, err := civogo.NewClient(token, "LON1")
	if err != nil {
		return err
	}
	id := client.GetAccountID()

	if len(id) == 0 {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedCloudAccountAuth,
			p.l.NewError(p.ctx, "Invalid user"),
		)
	}
	p.l.Print(p.ctx, "Recieved accountId", "userId", id)

	if err := p.store.WriteCredentials(consts.CloudCivo,
		&statefile.CredentialsDocument{
			InfraProvider: consts.CloudCivo,
			Civo:          &statefile.CredentialsCivo{Token: token},
		}); err != nil {
		return err
	}

	return nil
}

func (p *Provider) Name(resName string) providers.Cloud {

	if err := validation.IsValidName(p.ctx, p.l, resName); err != nil {
		p.l.Error("Resource Name", err.Error())
		return nil
	}
	p.chResName <- resName
	return p
}

func (p *Provider) Role(resRole consts.KsctlRole) providers.Cloud {

	if !validation.ValidateRole(resRole) {
		p.l.Error("invalidParameters",
			ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidKsctlRole,
				p.l.NewError(p.ctx, "invalid role", "role", resRole)).Error(),
		)
		return nil
	}

	p.chRole <- resRole
	p.l.Debug(p.ctx, "Printing", "Role", resRole)
	return p
}

func (p *Provider) VMType(size string) providers.Cloud {

	if err := isValidVMSize(p, size); err != nil {
		p.l.Error("VM", err.Error())
		return nil
	}
	p.chVMType <- size
	p.l.Debug(p.ctx, "Printing", "VMSize", size)
	return p
}

func (p *Provider) Visibility(toBePublic bool) providers.Cloud {
	p.public = toBePublic
	p.l.Debug(p.ctx, "Printing", "willBePublic", toBePublic)
	return p
}

func (p *Provider) SupportForApplications() bool {
	return true
}

func (p *Provider) aggregratedApps(s []string) (ret string) {
	if len(s) == 0 {
		ret = "traefik2-nodeport,metrics-server" // default: applications
	} else {
		ret = strings.Join(s, ",") + ",traefik2-nodeport,metrics-server"
	}
	p.l.Debug(p.ctx, "Printing", "apps", ret)
	return
}

func (p *Provider) Application(s []string) (externalApps bool) {
	p.apps = p.aggregratedApps(s)
	return false
}

func (p *Provider) CNI(s string) (externalCNI bool) {

	p.l.Debug(p.ctx, "Printing", "cni", s)
	switch consts.KsctlValidCNIPlugin(s) {
	case consts.CNICilium, consts.CNIFlannel:
		p.cni = s
	case "":
		p.cni = string(consts.CNIFlannel)
	default:
		// nothing external
		p.cni = string(consts.CNINone)
		return true
	}

	return false
}

func (p *Provider) k8sVersion(ver string) (string, error) {
	if len(ver) == 0 {
		return "1.26.4-k3s1", nil
	}

	ver = ver + "-k3s1"
	if err := isValidK8sVersion(p, ver); err != nil {
		return "", err
	}
	p.l.Debug(p.ctx, "Printing", "k8sVersion", ver)
	return ver, nil
}

func (p *Provider) ManagedK8sVersion(ver string) providers.Cloud {
	v, err := p.k8sVersion(ver)
	if err != nil {
		p.l.Error("Managed k8s version", err.Error())
		return nil
	}
	p.K8sVersion = v
	return p
}

func (p *Provider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice[string](p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
	p.l.Debug(p.ctx, "Printing", "hostnameOfWorkerPlanes", hostnames)
	return hostnames
}

func (p *Provider) NoOfControlPlane(no int, setter bool) (int, error) {
	p.l.Debug(p.ctx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfControlplane,
				p.l.NewError(p.ctx, "unable to fetch controlplane instanceIDs"),
			)
		}
		p.l.Debug(p.ctx, "Printing", "InstanceIDsOfControlplanes", p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		return len(p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoCP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		if currLen == 0 {
			p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs = make([]string, no)
			p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoControlPlanes.Hostnames = make([]string, no)
			p.state.CloudInfra.Civo.InfoControlPlanes.VMSizes = make([]string, no)
		}
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs", p.state.CloudInfra.Civo.InfoControlPlanes.VMIDs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs", p.state.CloudInfra.Civo.InfoControlPlanes.PublicIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs", p.state.CloudInfra.Civo.InfoControlPlanes.PrivateIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoControlPlanes.Hostnames", p.state.CloudInfra.Civo.InfoControlPlanes.Hostnames)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoControlPlanes.VMSizes", p.state.CloudInfra.Civo.InfoControlPlanes.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfControlplane,
		p.l.NewError(p.ctx, "constrains for no of controlplane >= 3 and odd number"),
	)
}

func (p *Provider) NoOfDataStore(no int, setter bool) (int, error) {
	p.l.Debug(p.ctx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Civo.InfoDatabase.VMIDs == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfDatastore,
				p.l.NewError(p.ctx, "unable to fetch DataStore instanceID"),
			)
		}

		p.l.Debug(p.ctx, "Printing", "InstanceIDsOfDatabaseNode", p.state.CloudInfra.Civo.InfoDatabase.VMIDs)

		return len(p.state.CloudInfra.Civo.InfoDatabase.VMIDs), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoDS = no

		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Civo.InfoDatabase.VMIDs)
		if currLen == 0 {
			p.state.CloudInfra.Civo.InfoDatabase.VMIDs = make([]string, no)
			p.state.CloudInfra.Civo.InfoDatabase.PublicIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoDatabase.Hostnames = make([]string, no)
			p.state.CloudInfra.Civo.InfoDatabase.VMSizes = make([]string, no)
		}

		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoDatabase.VMIDs", p.state.CloudInfra.Civo.InfoDatabase.VMIDs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoDatabase.PublicIPs", p.state.CloudInfra.Civo.InfoDatabase.PublicIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs", p.state.CloudInfra.Civo.InfoDatabase.PrivateIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoDatabase.Hostnames", p.state.CloudInfra.Civo.InfoDatabase.Hostnames)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoDatabase.VMSizes", p.state.CloudInfra.Civo.InfoDatabase.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfDatastore,
		p.l.NewError(p.ctx, "constrains for no of Datastore>= 3 and odd number"),
	)
}

func (p *Provider) NoOfWorkerPlane(no int, setter bool) (int, error) {
	p.l.Debug(p.ctx, "Printing", "desiredNumber", no, "setterOrNot", setter)

	if !setter {
		// delete operation
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfWorkerplane,
				p.l.NewError(p.ctx, "unable to fetch WorkerNode instanceIDs"),
			)
		}

		p.l.Debug(p.ctx, "Printing", "InstanceIDsOfWorkerPlane", p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		return len(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs), nil
	}
	if no >= 0 {
		p.NoWP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		currLen := len(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)

		newLen := no

		if currLen == 0 {
			p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = make([]string, no)
			p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = make([]string, no)
			p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = append(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs, "")
					p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = append(p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs, "")
					p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = append(p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs, "")
					p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = append(p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames, "")
					p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = append(p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes, "")
				}
			} else {
				// for downscaling
				p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs = p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs[:newLen]
				p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs = p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs[:newLen]
				p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs = p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs[:newLen]
				p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames = p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames[:newLen]
				p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes = p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}
		err := p.store.Write(p.state)
		if err != nil {
			return -1, err
		}

		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs", p.state.CloudInfra.Civo.InfoWorkerPlanes.VMIDs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs", p.state.CloudInfra.Civo.InfoWorkerPlanes.PublicIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs", p.state.CloudInfra.Civo.InfoWorkerPlanes.PrivateIPs)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames", p.state.CloudInfra.Civo.InfoWorkerPlanes.Hostnames)
		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes", p.state.CloudInfra.Civo.InfoWorkerPlanes.VMSizes)
		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfWorkerplane,
		p.l.NewError(p.ctx, "constrains for no of workerplane >= 0"),
	)
}

func (p *Provider) GetRAWClusterInfos() ([]logger.ClusterDataForLogging, error) {

	var data []logger.ClusterDataForLogging

	clusters, err := p.store.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudCivo),
		consts.ClusterType: "",
	})
	if err != nil {
		return nil, err
	}

	convertToAllClusterDataType := func(st *statefile.StorageDocument, r consts.KsctlRole) (v []logger.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Civo.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, logger.VMData{
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
				v = append(v, logger.VMData{
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
				v = append(v, logger.VMData{
					VMID:       o.VMIDs[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Civo.FirewallIDDatabaseNodes,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
				})
			}

		default:
			v = append(v, logger.VMData{
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
			data = append(data, logger.ClusterDataForLogging{
				CloudProvider: consts.CloudCivo,
				ClusterType:   K,
				Name:          v.ClusterName,
				Region:        v.Region,
				CP:            convertToAllClusterDataType(v, consts.RoleCp),
				WP:            convertToAllClusterDataType(v, consts.RoleWp),
				DS:            convertToAllClusterDataType(v, consts.RoleDs),
				LB:            convertToAllClusterDataType(v, consts.RoleLb)[0],

				Mgt: logger.VMData{
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
			p.l.Debug(p.ctx, "Printing", "cloudClusterInfoFetched", data)
		}
	}

	return data, nil
}

func (p *Provider) isPresent(cType consts.KsctlClusterType) error {
	err := p.store.AlreadyCreated(consts.CloudCivo, p.Region, p.ClusterName, cType)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) IsPresent() error {
	if p.IsHA {
		return p.isPresent(consts.ClusterTypeHa)
	}
	return p.isPresent(consts.ClusterTypeMang)
}

func (p *Provider) GetKubeconfig() (*string, error) {
	_read, err := p.store.Read()
	if err != nil {
		p.l.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig := _read.ClusterKubeConfig
	return &kubeconfig, nil
}
