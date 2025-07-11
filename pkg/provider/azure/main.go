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

package azure

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
	"github.com/ksctl/ksctl/v2/pkg/validation"
)

type Provider struct {
	l           logger.Logger
	ctx         context.Context
	ksctlConfig context.Context
	state       *statefile.StorageDocument
	store       storage.Storage
	mu          sync.Mutex

	controller.Metadata

	resourceGroup string

	public bool

	managedAddonCNI string
	managedAddonApp map[string]map[string]*string

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client CloudSDK

	tenantID       string
	subscriptionID string
	clientID       string
	clientSecret   string
}

func NewClient(
	ctx context.Context,
	l logger.Logger,
	ksctlConfig context.Context,
	meta controller.Metadata,
	state *statefile.StorageDocument,
	storage storage.Storage,
	ClientOption func() CloudSDK,
) (*Provider, error) {
	p := new(Provider)
	p.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, string(consts.CloudAzure))
	p.state = state
	p.Metadata = meta
	p.l = l
	p.client = ClientOption()
	p.store = storage

	p.ksctlConfig = ksctlConfig

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

func (p *Provider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames)
	p.l.Debug(p.ctx, "Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
}

func (p *Provider) ManagedK8sVersion(ver string) provider.Cloud {
	p.l.Debug(p.ctx, "Printing", "K8sVersion", ver)
	if err := p.isValidK8sVersion(ver); err != nil {
		p.l.Error("Managed k8s version", err.Error())
		return nil
	}

	p.K8sVersion = ver
	return p
}

func (p *Provider) GetStateForHACluster() (provider.CloudResourceState, error) {
	payload := provider.CloudResourceState{
		SSHPrivateKey: p.state.SSHKeyPair.PrivateKey,
		SSHUserName:   p.state.CloudInfra.Azure.B.SSHUser,
		ClusterName:   p.state.ClusterName,
		Provider:      p.state.InfraProvider,
		Region:        p.state.Region,
		ClusterType:   p.ClusterType,

		// public IPs
		IPv4ControlPlanes: utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  p.state.CloudInfra.Azure.InfoLoadBalancer.PublicIP,

		// Private IPs
		PrivateIPv4ControlPlanes: utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice(p.state.CloudInfra.Azure.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  p.state.CloudInfra.Azure.InfoLoadBalancer.PrivateIP,
	}

	p.l.Success(p.ctx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
}

func (p *Provider) InitState(operation consts.KsctlOperation) error {
	v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlAzureCredentials)
	if !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "missing azure credentials"),
		)
	}
	extractedCreds := statefile.CredentialsAzure{}
	if err := json.Unmarshal([]byte(v), &extractedCreds); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "failed to get azure credentials", "reason", err),
		)
	}

	p.tenantID = extractedCreds.TenantID
	p.subscriptionID = extractedCreds.SubscriptionID
	p.clientID = extractedCreds.ClientID
	p.clientSecret = extractedCreds.ClientSecret

	p.chResName = make(chan string, 1)
	p.chRole = make(chan consts.KsctlRole, 1)
	p.chVMType = make(chan string, 1)

	p.resourceGroup = generateResourceGroupName(p.ClusterName, string(p.ClusterType))

	errLoadState := p.loadStateHelper()

	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && p.state.CloudInfra.Azure.B.IsCompleted {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				p.l.NewError(p.ctx, "cluster already exist", "name", p.state.ClusterName, "region", p.state.Region),
			)
		}
		if errLoadState == nil && !p.state.CloudInfra.Azure.B.IsCompleted {
			p.l.Debug(p.ctx, "RESUME triggered!!")
		} else {
			p.l.Debug(p.ctx, "Fresh state!!")
			owner, team := "", ""

			if v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlContextUser); ok {
				owner = v
			}

			if v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlContextGroup); ok {
				team = v
			}

			p.state.PlatformSpec.Team = team
			p.state.PlatformSpec.Owner = owner
			p.state.PlatformSpec.State = statefile.Creating
			p.state.ClusterName = p.ClusterName
			p.state.Region = p.Region
			p.state.InfraProvider = consts.CloudAzure
			p.state.ClusterType = string(p.ClusterType)
			p.state.CloudInfra = &statefile.InfrastructureState{
				Azure: &statefile.StateConfigurationAzure{},
			}
		}

	case consts.OperationDelete:
		if errLoadState != nil {
			return errLoadState
		}
		p.l.Debug(p.ctx, "Delete resource(s)")

		p.state.PlatformSpec.State = statefile.Deleting

	case consts.OperationGet:
		if errLoadState != nil {
			return errLoadState
		}
		p.l.Debug(p.ctx, "Get storage")

	case consts.OperationConfigure, consts.OperationScale:
		if errLoadState != nil {
			return errLoadState
		}
		p.l.Debug(p.ctx, "Configuring resource(s)")

		p.state.PlatformSpec.State = statefile.Configuring

	default:
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidOperation,
			p.l.NewError(p.ctx, "Invalid operation for init state"),
		)
	}

	if operation != consts.OperationGet {
		if err := p.store.Write(p.state); err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				p.l.NewError(p.ctx, "failed to write the state", "Reason", err),
			)
		}
	}

	if err := p.client.InitClient(p); err != nil {
		return err
	}

	if err := p.validationOfArguments(); err != nil {
		return err
	}

	p.l.Debug(p.ctx, "init cloud state")

	return nil
}

func (p *Provider) Name(resName string) provider.Cloud {

	if err := validation.IsValidName(p.ctx, p.l, resName); err != nil {
		p.l.Error("Resource Name", err.Error())
		return nil
	}

	p.chResName <- resName
	return p
}

func (p *Provider) Role(resRole consts.KsctlRole) provider.Cloud {

	if !validation.ValidateRole(resRole) {
		p.l.Error("invalidParameters",
			ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidKsctlRole,
				p.l.NewError(p.ctx, "invalid role", "role", resRole)).
				Error(),
		)
		return nil
	}

	p.chRole <- resRole
	p.l.Debug(p.ctx, "Printing", "Role", resRole)
	return p
}

func (p *Provider) VMType(size string) provider.Cloud {

	if err := p.isValidVMSize(size); err != nil {
		p.l.Error("VM", err.Error())
		return nil
	}
	p.chVMType <- size

	return p
}

func (p *Provider) Visibility(toBePublic bool) provider.Cloud {
	p.public = toBePublic
	return p
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
		if p.state.CloudInfra.Azure.InfoControlPlanes.Names == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfControlplane,
				p.l.NewError(p.ctx, "unable to fetch controlplane instanceIDs"),
			)
		}

		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Azure.InfoControlPlanes.Names", p.state.CloudInfra.Azure.InfoControlPlanes.Names)
		return len(p.state.CloudInfra.Azure.InfoControlPlanes.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoCP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Azure.InfoControlPlanes.Names)
		if currLen == 0 {
			p.state.CloudInfra.Azure.InfoControlPlanes.Names = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.Hostnames = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.DiskNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.PublicIPIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoControlPlanes.VMSizes = make([]string, no)
		}

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
		if p.state.CloudInfra.Azure.InfoDatabase.Names == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfDatastore,
				p.l.NewError(p.ctx, "unable to fetch DataStore instanceID"),
			)
		}

		return len(p.state.CloudInfra.Azure.InfoDatabase.Names), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoDS = no

		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Azure.InfoDatabase.Names)
		if currLen == 0 {
			p.state.CloudInfra.Azure.InfoDatabase.Names = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.Hostnames = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.DiskNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.PublicIPIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoDatabase.VMSizes = make([]string, no)
		}

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
		if p.state.CloudInfra.Azure.InfoWorkerPlanes.Names == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfWorkerplane,
				p.l.NewError(p.ctx, "unable to fetch WorkerNode instanceIDs"),
			)
		}
		p.l.Debug(p.ctx, "Prnting", "p.state.CloudInfra.Azure.InfoWorkerPlanes.Names", p.state.CloudInfra.Azure.InfoWorkerPlanes.Names)
		return len(p.state.CloudInfra.Azure.InfoWorkerPlanes.Names), nil
	}
	if no >= 0 {
		p.NoWP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		currLen := len(p.state.CloudInfra.Azure.InfoWorkerPlanes.Names)

		newLen := no

		if currLen == 0 {
			p.state.CloudInfra.Azure.InfoWorkerPlanes.Names = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = make([]string, no)
			p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = make([]string, no)
		} else {
			if currLen == newLen {
				// no changes needed
				return -1, nil
			} else if currLen < newLen {
				// for up-scaling
				for i := currLen; i < newLen; i++ {
					p.state.CloudInfra.Azure.InfoWorkerPlanes.Names = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.Names, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs, "")
					p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = append(p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes, "")
				}
			} else {
				// for downscaling
				p.state.CloudInfra.Azure.InfoWorkerPlanes.Names = p.state.CloudInfra.Azure.InfoWorkerPlanes.Names[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames = p.state.CloudInfra.Azure.InfoWorkerPlanes.Hostnames[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPs[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs = p.state.CloudInfra.Azure.InfoWorkerPlanes.PrivateIPs[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames = p.state.CloudInfra.Azure.InfoWorkerPlanes.DiskNames[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceNames[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs = p.state.CloudInfra.Azure.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPNames[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs = p.state.CloudInfra.Azure.InfoWorkerPlanes.PublicIPIDs[:newLen]
				p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes = p.state.CloudInfra.Azure.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}

		if err := p.store.Write(p.state); err != nil {
			return -1, err
		}

		p.l.Debug(p.ctx, "Printing", "p.state.CloudInfra.Azure.InfoWorkerPlanes", p.state.CloudInfra.Azure.InfoWorkerPlanes)

		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfWorkerplane,
		p.l.NewError(p.ctx, "constrains for no of workerplane >= 0"),
	)
}

func (p *Provider) GetRAWClusterInfos() ([]provider.ClusterData, error) {

	var data []provider.ClusterData

	clusterType := ""
	if p.ClusterType != "" {
		clusterType = string(p.ClusterType)
	}

	clusters, err := p.store.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAzure),
		consts.ClusterType: clusterType,
	})
	if err != nil {
		return nil, err
	}

	convertToAllClusterDataType := func(st *statefile.StorageDocument, r consts.KsctlRole) (v []provider.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Azure.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, provider.VMData{
					Name:         o.Names[i],
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
				v = append(v, provider.VMData{
					Name:         o.Names[i],
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
				v = append(v, provider.VMData{
					Name:         o.Names[i],
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
			v = append(v, provider.VMData{
				Name:         st.CloudInfra.Azure.InfoLoadBalancer.Name,
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
			data = append(data, provider.ClusterData{
				Owner:         v.PlatformSpec.Owner,
				Team:          v.PlatformSpec.Team,
				State:         v.PlatformSpec.State,
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
				Mgt: provider.VMData{
					VMSize: v.CloudInfra.Azure.ManagedNodeSize,
				},
				ManagedK8sName:  v.CloudInfra.Azure.ManagedClusterName,
				NetworkName:     v.CloudInfra.Azure.VirtualNetworkName,
				NetworkID:       v.CloudInfra.Azure.VirtualNetworkID,
				ResourceGrpName: v.CloudInfra.Azure.ResourceGroupName,
				SSHKeyName:      v.CloudInfra.Azure.B.SSHKeyName,
				SSHKeyID:        v.CloudInfra.Azure.B.SSHID,

				K8sDistro: v.BootstrapProvider,
				HAProxyVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeSelfMang) {
						if v.Versions.HAProxy != nil {
							return *v.Versions.HAProxy
						}
						return ""
					}
					return ""
				}(),
				EtcdVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeSelfMang) {
						if v.Versions.Etcd != nil {
							return *v.Versions.Etcd
						}
						return ""
					}
					return ""
				}(),
				K8sVersion: func() string {
					switch v.BootstrapProvider {
					case consts.K8sK3s:
						if v.Versions.K3s == nil {
							return ""
						}
						return *v.Versions.K3s
					case consts.K8sKubeadm:
						if v.Versions.Kubeadm == nil {
							return ""
						}
						return *v.Versions.Kubeadm
					default:
						if v.Versions.Aks == nil {
							return ""
						}
						return *v.Versions.Aks
					}
				}(),
				Apps: func() (_a []string) {
					for _, a := range v.ProvisionerAddons.Apps {
						_a = append(_a, a.String())
					}
					return
				}(),
				Cni: v.ProvisionerAddons.Cni.String(),
			})

		}
	}

	return data, nil
}

func (p *Provider) isPresent(cType consts.KsctlClusterType) error {
	err := p.store.AlreadyCreated(consts.CloudAzure, p.Region, p.ClusterName, cType)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) IsPresent() error {
	return p.isPresent(p.ClusterType)
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
