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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"
)

func (obj *AwsProvider) NewFirewall(storage ksctlTypes.StorageFactory) error {

	role := <-obj.chRole
	name := <-obj.chResName
	_, err := obj.CreateSecurityGroup(name, role)
	if err != nil {
		return err
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	return nil
}

func (obj *AwsProvider) DelFirewall(storage ksctlTypes.StorageFactory) error {

	role := <-obj.chRole

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs
	case consts.RoleWp:
		nsg = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs
	case consts.RoleLb:
		nsg = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID
	case consts.RoleDs:
		nsg = mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs
	}

	if len(nsg) == 0 {
		log.Print(awsCtx, "skipped firewall already deleted")
		return nil
	} else {
		err := obj.client.BeginDeleteSecurityGrp(awsCtx, nsg)
		if err != nil {
			return err
		}

		switch role {
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs = ""
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs = ""
		}

		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}

		log.Success(awsCtx, "Deleted the security group", "id", nsg)
	}

	return nil
}

func (obj *AwsProvider) CreateSecurityGroup(name string, role consts.KsctlRole) (string, error) {

	SecurityGroupInput := ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(name + "-" + string(role)),
		VpcId:       aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
	}

	switch role {
	case consts.RoleCp:
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs != "" {
			log.Success(awsCtx, "skipped already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs)
			return mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs, nil
		}
	case consts.RoleWp:
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs != "" {
			log.Success(awsCtx, "skipped already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs)
			return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs, nil
		}
	case consts.RoleLb:
		if mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID != "" {
			log.Success(awsCtx, "skipped already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID)
			return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID, nil
		}
	case consts.RoleDs:
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs != "" {
			log.Success(awsCtx, "skipped already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs)
			return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs, nil
		}
	}

	kubernetesDistro := mainStateDocument.BootstrapProvider
	netCidr := mainStateDocument.CloudInfra.Aws.VpcCidr

	SecurityGroup, err := obj.client.BeginCreateSecurityGroup(awsCtx, SecurityGroupInput)
	if err != nil {
		return "", err
	}

	err = obj.createSecurityGroupRules(
		consts.KsctlKubernetes(kubernetesDistro),
		netCidr, role, SecurityGroup)
	if err != nil {
		return "", err
	}

	log.Success(awsCtx, "Created SecurityGroup", "name", name)

	return *SecurityGroup.GroupId, nil
}

func (obj *AwsProvider) createSecurityGroupRules(
	bootstrap consts.KsctlKubernetes,
	netCidr string,
	role consts.KsctlRole,
	SecurityGroup *ec2.CreateSecurityGroupOutput,
) (err error) {

	var ingressrules ec2.AuthorizeSecurityGroupIngressInput
	var egressrules ec2.AuthorizeSecurityGroupEgressInput

	switch role {
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleLoadBalancer(SecurityGroup.GroupId)

	case consts.RoleCp:
		mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleControlPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleWp:
		mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleWorkerPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleDs:
		mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleDataStore(
			SecurityGroup.GroupId,
			netCidr,
		)

	}

	if err := obj.client.AuthorizeSecurityGroupIngress(awsCtx, ingressrules); err != nil {
		return err
	}

	if err := obj.client.AuthorizeSecurityGroupEgress(awsCtx, egressrules); err != nil {
		return err
	}
	return nil
}

func convertToProviderSpecific(_rules []helpers.FirewallRule, SgId *string) (ec2.AuthorizeSecurityGroupIngressInput, ec2.AuthorizeSecurityGroupEgressInput) {

	ingressRules := []types.IpPermission{}
	egressRules := []types.IpPermission{}

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
		helpers.FirewallForControlplane_BASE(internalNetCidr, bootstrap),
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
		helpers.FirewallForWorkerplane_BASE(internalNetCidr, bootstrap),
		sgid,
	)
}

func firewallRuleLoadBalancer(
	sgid *string,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {
	return convertToProviderSpecific(
		helpers.FirewallForLoadBalancer_BASE(),
		sgid,
	)
}

func firewallRuleDataStore(
	sgid *string,
	internalNetCidr string,
) (ec2.AuthorizeSecurityGroupIngressInput,
	ec2.AuthorizeSecurityGroupEgressInput) {
	return convertToProviderSpecific(
		helpers.FirewallForDataStore_BASE(internalNetCidr),
		sgid,
	)
}
