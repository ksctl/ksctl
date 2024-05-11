package aws

import (
	"context"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
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
		return log.NewError("Error saving state", "error", err)
	}

	return nil
}

func (obj *AwsProvider) DelFirewall(storage ksctlTypes.StorageFactory) error {

	role := <-obj.chRole

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup
	case consts.RoleWp:
		nsg = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup
	case consts.RoleLb:
		nsg = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup
	case consts.RoleDs:
		nsg = mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup
	default:
		return log.NewError("invalid role")
	}

	if len(nsg) == 0 {
		log.Print("skipped firewall already deleted")
		return nil
	} else {
		err := obj.client.BeginDeleteSecurityGrp(context.Background(), nsg)
		if err != nil {
			return err
		}

		switch role {
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup = ""
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup = ""
		default:
			return log.NewError("invalid role")
		}

		err = storage.Write(mainStateDocument)
		if err != nil {
			return log.NewError("Error saving state", "error", err)
		}

		log.Success("Deleted the security group", "id", nsg)
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
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleWp:
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleLb:
		if mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup, nil
		}
	case consts.RoleDs:
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group", "id", mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup, nil
		}
	}

	kubernetesDistro := mainStateDocument.CloudInfra.Aws.B.KubernetesDistro
	netCidr := mainStateDocument.CloudInfra.Aws.VpcCidr

	SecurityGroup, err := obj.client.BeginCreateSecurityGroup(context.Background(), SecurityGroupInput)
	if err != nil {
		return "", err
	}

	err = obj.createSecurityGroupRules(
		consts.KsctlKubernetes(kubernetesDistro),
		netCidr, role, SecurityGroup)
	if err != nil {
		return "", err
	}

	log.Success("Created SecurityGroup", "name", name)

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
		mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleLoadBalancer(SecurityGroup.GroupId)

	case consts.RoleCp:
		mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleControlPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleWp:
		mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleWorkerPlane(
			SecurityGroup.GroupId,
			netCidr,
			bootstrap,
		)

	case consts.RoleDs:
		mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup = *SecurityGroup.GroupId
		ingressrules, egressrules = firewallRuleDataStore(
			SecurityGroup.GroupId,
			netCidr,
		)

	default:
		return log.NewError("Error creating security group", "error", "invalid role")
	}

	if err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), ingressrules); err != nil {
		return err
	}

	if err := obj.client.AuthorizeSecurityGroupEgress(context.Background(), egressrules); err != nil {
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
			FromPort:   to.Ptr[int32](int32(_startPort)),
			ToPort:     to.Ptr[int32](int32(_endPort)),
			IpProtocol: to.Ptr[string](protocol),
			IpRanges: []types.IpRange{
				{
					CidrIp:      to.Ptr[string](_r.Cidr),
					Description: to.Ptr[string](_r.Description),
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
