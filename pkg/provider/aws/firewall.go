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
	"github.com/ksctl/ksctl/pkg/firewall"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func (p *Provider) NewFirewall() error {

	role := <-p.chRole
	name := <-p.chResName
	_, err := p.CreateSecurityGroup(name, role)
	if err != nil {
		return err
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	return nil
}

func (p *Provider) DelFirewall() error {

	role := <-p.chRole

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs
	case consts.RoleWp:
		nsg = p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs
	case consts.RoleLb:
		nsg = p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID
	case consts.RoleDs:
		nsg = p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs
	}

	if len(nsg) == 0 {
		p.l.Print(p.ctx, "skipped firewall already deleted")
		return nil
	} else {
		err := p.client.BeginDeleteSecurityGrp(p.ctx, nsg)
		if err != nil {
			return err
		}

		switch role {
		case consts.RoleCp:
			p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs = ""
		case consts.RoleWp:
			p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs = ""
		case consts.RoleLb:
			p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID = ""
		case consts.RoleDs:
			p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs = ""
		}

		err = p.store.Write(p.state)
		if err != nil {
			return err
		}

		p.l.Success(p.ctx, "Deleted the security group", "id", nsg)
	}

	return nil
}

func (p *Provider) CreateSecurityGroup(name string, role consts.KsctlRole) (string, error) {

	SecurityGroupInput := ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(name + "-" + string(role)),
		VpcId:       aws.String(p.state.CloudInfra.Aws.VpcId),
	}

	switch role {
	case consts.RoleCp:
		if p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs != "" {
			p.l.Success(p.ctx, "skipped already created the security group", "id", p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs)
			return p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs, nil
		}
	case consts.RoleWp:
		if p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs != "" {
			p.l.Success(p.ctx, "skipped already created the security group", "id", p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs)
			return p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs, nil
		}
	case consts.RoleLb:
		if p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID != "" {
			p.l.Success(p.ctx, "skipped already created the security group", "id", p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID)
			return p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID, nil
		}
	case consts.RoleDs:
		if p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs != "" {
			p.l.Success(p.ctx, "skipped already created the security group", "id", p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs)
			return p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs, nil
		}
	}

	kubernetesDistro := p.state.BootstrapProvider
	netCidr := p.state.CloudInfra.Aws.VpcCidr

	SecurityGroup, err := p.client.BeginCreateSecurityGroup(p.ctx, SecurityGroupInput)
	if err != nil {
		return "", err
	}

	err = p.createSecurityGroupRules(
		kubernetesDistro,
		netCidr, role, SecurityGroup)
	if err != nil {
		return "", err
	}

	p.l.Success(p.ctx, "Created SecurityGroup", "name", name)

	return *SecurityGroup.GroupId, nil
}

func (p *Provider) createSecurityGroupRules(
	bootstrap consts.KsctlKubernetes,
	netCidr string,
	role consts.KsctlRole,
	SecurityGroup *ec2.CreateSecurityGroupOutput,
) (err error) {

	var ingressrules ec2.AuthorizeSecurityGroupIngressInput
	var egressrules ec2.AuthorizeSecurityGroupEgressInput

	switch role {
	case consts.RoleLb:
		p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleLoadBalancer(SecurityGroup.GroupId)

	case consts.RoleCp:
		p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleControlPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleWp:
		p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleWorkerPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleDs:
		p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleDataStore(
			SecurityGroup.GroupId,
			netCidr,
		)

	}

	if err := p.client.AuthorizeSecurityGroupIngress(p.ctx, ingressrules); err != nil {
		return err
	}

	if err := p.client.AuthorizeSecurityGroupEgress(p.ctx, egressrules); err != nil {
		return err
	}
	return nil
}

func convertToProviderSpecific(_rules []firewall.FirewallRule, SgId *string) (ec2.AuthorizeSecurityGroupIngressInput, ec2.AuthorizeSecurityGroupEgressInput) {

	var ingressRules []types.IpPermission
	var egressRules []types.IpPermission

	for _, _r := range _rules {

		var protocol string

		switch _r.Protocol {
		case consts.FirewallActionTCP:
			protocol = "tcp"
		case consts.FirewallActionUDP:
			protocol = "udp"
		default:
			protocol = "tcp"
		}

		_startPort, _ := strconv.Atoi(_r.StartPort)
		_endPort, _ := strconv.Atoi(_r.EndPort)

		v := types.IpPermission{
			FromPort:   utilities.Ptr[int32](int32(_startPort)),
			ToPort:     utilities.Ptr[int32](int32(_endPort)),
			IpProtocol: utilities.Ptr[string](protocol),
			IpRanges: []types.IpRange{
				{
					CidrIp:      utilities.Ptr[string](_r.Cidr),
					Description: utilities.Ptr[string](_r.Description),
				},
			},
		}

		switch _r.Direction {
		case consts.FirewallActionIngress:
			ingressRules = append(ingressRules, v)
		case consts.FirewallActionEgress:
			egressRules = append(egressRules, v)
		}
	}

	return ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       SgId,
			IpPermissions: ingressRules,
		}, ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       SgId,
			IpPermissions: egressRules,
		}
}

func firewallRuleControlPlane(sgid *string,
	internalNetCidr string,
	bootstrap consts.KsctlKubernetes,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {

	return convertToProviderSpecific(
		firewall.FirewallforcontrolplaneBase(internalNetCidr, bootstrap),
		sgid,
	)
}

func firewallRuleWorkerPlane(
	sgid *string,
	internalNetCidr string,
	bootstrap consts.KsctlKubernetes,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {

	return convertToProviderSpecific(
		firewall.FirewallforworkerplaneBase(internalNetCidr, bootstrap),
		sgid,
	)
}

func firewallRuleLoadBalancer(
	sgid *string,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {
	return convertToProviderSpecific(
		firewall.FirewallforloadbalancerBase(),
		sgid,
	)
}

func firewallRuleDataStore(
	sgid *string,
	internalNetCidr string,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {
	return convertToProviderSpecific(
		firewall.FirewallfordatastoreBase(internalNetCidr),
		sgid,
	)
}
