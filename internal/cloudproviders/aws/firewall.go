package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func (obj *AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	role := <-obj.chRole
	<-obj.chResName
	_, err := obj.CreateSecurityGroup(role)
	if err != nil {
		return err
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError("Error saving state", "error", err)
	}

	return nil
}

func (obj *AwsProvider) DelFirewall(storage resources.StorageFactory) error {

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

		log.Success("deleted the security group ", "id", nsg)
	}

	return nil
}

func (obj *AwsProvider) CreateSecurityGroup(Role consts.KsctlRole) (string, error) {

	SecurityGroupInput := ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(string(Role + "securitygroup")),
		Description: aws.String(obj.clusterName + "securitygroup"),
		VpcId:       aws.String(mainStateDocument.CloudInfra.Aws.VpcId),
	}

	switch Role {
	case consts.RoleCp:
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleWp:
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id", mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleLb:
		if mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id", mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup, nil
		}
	case consts.RoleDs:
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id", mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup, nil
		}
	}

	SecurityGroup, err := obj.client.BeginCreateSecurityGroup(context.Background(), SecurityGroupInput)
	if err != nil {
		return "", err
	}

	err = obj.createSecurityGroupRules(Role, SecurityGroup)
	if err != nil {
		return "", err
	}

	return *SecurityGroup.GroupId, nil
}

// Anchor for next steps
func (obj *AwsProvider) createSecurityGroupRules(role consts.KsctlRole, SecurityGroup *ec2.CreateSecurityGroupOutput) (err error) {
	var ip_protocol string
	var cidr_ip string
	var from_port int32
	var to_port int32

	ip_protocol = "-1"
	cidr_ip = "0.0.0.0/0"
	from_port = 0
	to_port = 65535

	switch role {
	case consts.RoleLb:
		mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleWp:
		mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleDs:
		mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup = *SecurityGroup.GroupId
	default:
		return log.NewError("Error creating security group", "error", "invalid role")
	}
	ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    SecurityGroup.GroupId,
		IpProtocol: aws.String(ip_protocol),
		CidrIp:     aws.String(cidr_ip),
	}

	if err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), ingressrules); err != nil {
		return err
	}

	egressrules := ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: SecurityGroup.GroupId,
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String(ip_protocol),
				FromPort:   aws.Int32(from_port),
				ToPort:     aws.Int32(to_port),
			},
		},
	}

	if err := obj.client.AuthorizeSecurityGroupEgress(context.Background(), egressrules); err != nil {
		return err
	}
	return nil
}
