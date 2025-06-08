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

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/validation"

	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/utilities"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type StsPresignClientInteface interface {
	PresignGetCallerIdentity(ctx context.Context, input *sts.GetCallerIdentityInput, options ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type STSTokenRetriever struct {
	PresignClient StsPresignClientInteface
}

type customHTTPPresignerV4 struct {
	client  sts.HTTPPresignerV4
	headers map[string]string
}

type KubeConfigData struct {
	ClusterEndpoint          string
	CertificateAuthorityData string
	ClusterName              string
	Token                    string
}

type Provider struct {
	l           logger.Logger
	ctx         context.Context
	ksctlConfig context.Context
	state       *statefile.StorageDocument
	store       storage.Storage
	mu          sync.Mutex

	public bool

	controller.Metadata

	vpc string

	managedAddonCNI string
	managedAddonApp []string

	chResName chan string
	chRole    chan consts.KsctlRole
	chVMType  chan string

	client CloudSDK

	accessKeyId string
	secretKey   string
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
	p.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, string(consts.CloudAws))
	p.state = state
	p.Metadata = meta
	p.l = l
	p.client = ClientOption()
	p.store = storage

	p.ksctlConfig = ksctlConfig

	return p, nil
}

func (p *Provider) isPresent(cType consts.KsctlClusterType) error {
	err := p.store.AlreadyCreated(consts.CloudAws, p.Region, p.ClusterName, cType)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) IsPresent() error {
	return p.isPresent(p.ClusterType)
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

func (p *Provider) Name(resName string) provider.Cloud {

	if err := validation.IsValidName(p.ctx, p.l, resName); err != nil {
		p.l.Error(err.Error())
		return nil
	}
	p.chResName <- resName
	return p
}

func (p *Provider) InitState(operation consts.KsctlOperation) error {
	v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlAwsCredentials)
	if !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "missing aws credentials"),
		)
	}
	extractedCreds := statefile.CredentialsAws{}
	if err := json.Unmarshal([]byte(v), &extractedCreds); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			p.l.NewError(p.ctx, "failed to get aws credentials", "reason", err),
		)
	}

	p.accessKeyId = extractedCreds.AccessKeyId
	p.secretKey = extractedCreds.SecretAccessKey

	p.chResName = make(chan string, 1)
	p.chRole = make(chan consts.KsctlRole, 1)
	p.chVMType = make(chan string, 1)

	p.vpc = fmt.Sprintf("%s-ksctl-%s-vpc", p.ClusterName, p.ClusterType)

	errLoadState := p.loadStateHelper()

	switch operation {
	case consts.OperationCreate:
		if errLoadState == nil && p.state.CloudInfra.Aws.IsCompleted {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				p.l.NewError(p.ctx, "cluster already exist", "name", p.state.ClusterName, "region", p.state.Region),
			)
		}
		if errLoadState == nil && !p.state.CloudInfra.Aws.IsCompleted {
			p.l.Note(p.ctx, "Cluster state found but not completed, resuming operation")
		} else {
			p.l.Debug(p.ctx, "Fresh state!!")
			owner, team := "", ""

			if v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlContextUser); ok {
				owner = v
			}

			if v, ok := config.IsContextPresent(p.ksctlConfig, consts.KsctlContextTeam); ok {
				team = v
			}

			p.state = statefile.NewStorageDocument(
				p.ClusterName,
				p.Region,
				consts.CloudAws,
				p.ClusterType,
				team,
				owner,
			)

			p.state.PlatformSpec.State = statefile.Creating

			p.state.CloudInfra = &statefile.InfrastructureState{
				Aws: &statefile.StateConfigurationAws{},
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

func (p *Provider) GetStateForHACluster() (provider.CloudResourceState, error) {

	payload := provider.CloudResourceState{
		SSHPrivateKey:     p.state.SSHKeyPair.PrivateKey,
		SSHUserName:       p.state.CloudInfra.Aws.B.SSHUser,
		ClusterName:       p.state.ClusterName,
		Provider:          p.state.InfraProvider,
		Region:            p.state.Region,
		ClusterType:       p.ClusterType,
		IPv4ControlPlanes: utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs),
		IPv4DataStores:    utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoDatabase.PublicIPs),
		IPv4WorkerPlanes:  utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs),
		IPv4LoadBalancer:  p.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP,

		PrivateIPv4ControlPlanes: utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoControlPlanes.PrivateIPs),
		PrivateIPv4DataStores:    utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoDatabase.PrivateIPs),
		PrivateIPv4LoadBalancer:  p.state.CloudInfra.Aws.InfoLoadBalancer.PrivateIP,
	}

	p.l.Success(p.ctx, "Transferred Data, it's ready to be shipped!")
	return payload, nil
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

func (p *Provider) NoOfWorkerPlane(no int, setter bool) (int, error) {
	p.l.Debug(p.ctx, "Printing", "desiredNumber", no, "setterOrNot", setter)
	if !setter {
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfWorkerplane,
				p.l.NewError(p.ctx, "unable to fetch WorkerNode instanceIDs"),
			)
		}
		return len(p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames), nil
	}
	if no >= 0 {
		p.NoWP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		currLen := len(p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames)

		newLen := no

		if currLen == 0 {
			p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames = make([]string, no)
			p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = make([]string, no)
			p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = make([]string, no)
		} else {
			if currLen == newLen {
				return -1, nil
			} else if currLen < newLen {
				for i := currLen; i < newLen; i++ {
					p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames, "")
					p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds, "")
					p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs, "")
					p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs, "")
					p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs, "")
					p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = append(p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes, "")
				}
			} else {
				p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames = p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames[:newLen]
				p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds = p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[:newLen]
				p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs = p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[:newLen]
				p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs = p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[:newLen]
				p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs = p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[:newLen]
				p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes = p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[:newLen]
			}
		}

		if err := p.store.Write(p.state); err != nil {
			return -1, err
		}

		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfWorkerplane,
		p.l.NewError(p.ctx, "constrains for no of workerplane >= 0"),
	)

}

func (p *Provider) NoOfControlPlane(no int, setter bool) (int, error) {
	if !setter {
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Aws.InfoControlPlanes.HostNames == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfControlplane,
				p.l.NewError(p.ctx, "unable to fetch Controlplane instanceIDs"),
			)
		}

		return len(p.state.CloudInfra.Aws.InfoControlPlanes.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoCP = no
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Aws.InfoControlPlanes.HostNames)
		if currLen == 0 {
			p.state.CloudInfra.Aws.InfoControlPlanes.HostNames = make([]string, no)
			p.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds = make([]string, no)
			p.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoControlPlanes.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Aws.InfoControlPlanes.VMSizes = make([]string, no)
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
		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}
		if p.state.CloudInfra.Aws.InfoDatabase.HostNames == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidNoOfDatastore,
				p.l.NewError(p.ctx, "unable to fetch Datastore instanceIDs"),
			)
		}

		return len(p.state.CloudInfra.Aws.InfoDatabase.HostNames), nil
	}
	if no >= 3 && (no&1) == 1 {
		p.NoDS = no

		if p.state == nil {
			return -1, ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidOperation,
				p.l.NewError(p.ctx, "state init not called!"),
			)
		}

		currLen := len(p.state.CloudInfra.Aws.InfoDatabase.HostNames)
		if currLen == 0 {
			p.state.CloudInfra.Aws.InfoDatabase.HostNames = make([]string, no)
			p.state.CloudInfra.Aws.InfoDatabase.InstanceIds = make([]string, no)
			p.state.CloudInfra.Aws.InfoDatabase.PublicIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoDatabase.PrivateIPs = make([]string, no)
			p.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs = make([]string, no)
			p.state.CloudInfra.Aws.InfoDatabase.VMSizes = make([]string, no)
		}

		return -1, nil
	}
	return -1, ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidNoOfDatastore,
		p.l.NewError(p.ctx, "constrains for no of Datastore>= 3 and odd number"),
	)
}

func (p *Provider) GetHostNameAllWorkerNode() []string {
	hostnames := utilities.DeepCopySlice(p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames)
	p.l.Debug(p.ctx, "Printing", "hostnameWorkerPlanes", hostnames)
	return hostnames
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

func (p *Provider) GetRAWClusterInfos() ([]provider.ClusterData, error) {

	var data []provider.ClusterData

	clusterType := ""
	if p.ClusterType != "" {
		clusterType = string(p.ClusterType)
	}

	clusters, err := p.store.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       string(consts.CloudAws),
		consts.ClusterType: clusterType,
	})
	if err != nil {
		return nil, err
	}

	subnetEvaluator := func(v []string) string {
		if len(v) > 0 {
			return v[0]
		}
		return ""
	}

	convertToAllClusterDataType := func(st *statefile.StorageDocument, r consts.KsctlRole) (v []provider.VMData) {

		switch r {
		case consts.RoleCp:
			o := st.CloudInfra.Aws.InfoControlPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, provider.VMData{
					Id:         o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   subnetEvaluator(st.CloudInfra.Aws.SubnetIDs),
					SubnetName: subnetEvaluator(st.CloudInfra.Aws.SubnetNames),
				})
			}

		case consts.RoleWp:
			o := st.CloudInfra.Aws.InfoWorkerPlanes
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, provider.VMData{
					Id:         o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   subnetEvaluator(st.CloudInfra.Aws.SubnetIDs),
					SubnetName: subnetEvaluator(st.CloudInfra.Aws.SubnetNames),
				})
			}

		case consts.RoleDs:
			o := st.CloudInfra.Aws.InfoDatabase
			no := len(o.VMSizes)
			for i := 0; i < no; i++ {
				v = append(v, provider.VMData{
					Id:         o.InstanceIds[i],
					VMSize:     o.VMSizes[i],
					FirewallID: st.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs,
					PublicIP:   o.PublicIPs[i],
					PrivateIP:  o.PrivateIPs[i],
					SubnetID:   subnetEvaluator(st.CloudInfra.Aws.SubnetIDs),
					SubnetName: subnetEvaluator(st.CloudInfra.Aws.SubnetNames),
				})
			}

		default:
			v = append(v, provider.VMData{
				Id:         st.CloudInfra.Aws.InfoLoadBalancer.InstanceID,
				VMSize:     st.CloudInfra.Aws.InfoLoadBalancer.VMSize,
				FirewallID: st.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID,
				PublicIP:   st.CloudInfra.Aws.InfoLoadBalancer.PublicIP,
				PrivateIP:  st.CloudInfra.Aws.InfoLoadBalancer.PrivateIP,
				SubnetID:   subnetEvaluator(st.CloudInfra.Aws.SubnetIDs),
				SubnetName: subnetEvaluator(st.CloudInfra.Aws.SubnetNames),
			})
		}
		return v
	}

	for K, Vs := range clusters {
		for _, v := range Vs {
			data = append(data, provider.ClusterData{
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
				Mgt: provider.VMData{
					VMSize: v.CloudInfra.Aws.ManagedNodeSize,
				},
				ManagedK8sID: v.CloudInfra.Aws.ManagedClusterArn,
				NetworkID:    v.CloudInfra.Aws.VpcId,
				NetworkName:  v.CloudInfra.Aws.VpcName,
				SSHKeyID:     v.CloudInfra.Aws.B.SSHID,
				SSHKeyName:   v.CloudInfra.Aws.B.SSHKeyName,

				K8sDistro: v.BootstrapProvider,
				HAProxyVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeSelfMang) {
						if v.Versions.HAProxy == nil {
							return ""
						}
						return *v.Versions.HAProxy
					}
					return ""
				}(),
				EtcdVersion: func() string {
					if v.ClusterType == string(consts.ClusterTypeSelfMang) {
						if v.Versions.Etcd == nil {
							return ""
						}
						return *v.Versions.Etcd
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
						if v.Versions.Eks == nil {
							return ""
						}
						return *v.Versions.Eks
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

func (p *Provider) GetKubeconfig() (*string, error) {

	if p.ClusterType == consts.ClusterTypeMang {
		kubeconfig, err := p.client.GetKubeConfig(p.ctx, p.state.CloudInfra.Aws.ManagedClusterName)
		if err != nil {
			return nil, err
		}
		return &kubeconfig, nil
	}

	kubeconfig := p.state.ClusterKubeConfig
	return &kubeconfig, nil
}
