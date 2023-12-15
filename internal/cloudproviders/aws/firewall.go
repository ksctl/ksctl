package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/kubesimplify/ksctl/pkg/resources"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

func (obj *AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	_, err := obj.CreateSecurityGroup(role)
	if err != nil {
		fmt.Println(err)
	}

	if err := saveStateHelper(storage); err != nil {
		log.Error("Error saving state", "error", err)
	}

	return nil
}

func (obj *AwsProvider) DelFirewall(factory resources.StorageFactory) error {
	role := obj.metadata.role
	obj.mxRole.Unlock()

	nsg := ""
	switch role {
	case consts.RoleCp:
		nsg = awsCloudState.InfoControlPlanes.NetworkSecurityGroup
	case consts.RoleWp:
		nsg = awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup
	case consts.RoleLb:
		nsg = awsCloudState.InfoLoadBalancer.NetworkSecurityGroup
	case consts.RoleDs:
		nsg = awsCloudState.InfoDatabase.NetworkSecurityGroup
	default:
		return fmt.Errorf("invalid role")
	}

	if len(nsg) == 0 {
		log.Print("skipped firewall already deleted")
		return nil
	} else {
		err := obj.client.BeginDeleteSecurityGrp(context.Background(), obj.ec2Client(), nsg)
		if err != nil {
			log.Error("Error deleting security group", "error", err)
		}

		switch role {
		case consts.RoleCp:
			awsCloudState.InfoControlPlanes.NetworkSecurityGroup = ""
		case consts.RoleWp:
			awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup = ""
		case consts.RoleLb:
			awsCloudState.InfoLoadBalancer.NetworkSecurityGroup = ""
		case consts.RoleDs:
			awsCloudState.InfoDatabase.NetworkSecurityGroup = ""
		default:
			return fmt.Errorf("invalid role")
		}

		err = saveStateHelper(factory)
		if err != nil {
			log.Error("Error saving state", "error", err)
		}

		log.Success("[aws] deleted the security group ", "id: ", nsg)
	}

	return nil
}

func (obj *AwsProvider) CreateSecurityGroup(Role consts.KsctlRole) (string, error) {

	SecurityGroupInput := ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(string(Role + "securitygroup")),
		Description: aws.String(obj.clusterName + "securitygroup"),
		VpcId:       aws.String(awsCloudState.VPCID),
	}

	switch Role {
	case consts.RoleCp:
		if awsCloudState.InfoControlPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", awsCloudState.InfoControlPlanes.NetworkSecurityGroup)
			return awsCloudState.InfoControlPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleWp:
		if awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup)
			return awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleLb:
		if awsCloudState.InfoLoadBalancer.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", awsCloudState.InfoLoadBalancer.NetworkSecurityGroup)
			return awsCloudState.InfoLoadBalancer.NetworkSecurityGroup, nil
		}
	case consts.RoleDs:
		if awsCloudState.InfoDatabase.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", awsCloudState.InfoDatabase.NetworkSecurityGroup)
			return awsCloudState.InfoDatabase.NetworkSecurityGroup, nil
		}
	}

	SecurityGroup, err := obj.client.BeginCreateSecurityGroup(context.Background(), obj.ec2Client(), SecurityGroupInput)
	if err != nil {
		fmt.Println(err)
	}

	switch Role {
	case consts.RoleCp:
		awsCloudState.InfoControlPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId

		func() {
			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules)

			egressrules := ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: SecurityGroup.GroupId,
				IpPermissions: []types.IpPermission{
					{
						IpProtocol: aws.String("-1"),
						FromPort:   aws.Int32(0),
						ToPort:     aws.Int32(65535),
					},
				},
			}

			obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules)
		}()

	case consts.RoleWp:
		func() {

			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules)
			egressrules := ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: SecurityGroup.GroupId,
				IpPermissions: []types.IpPermission{
					{
						IpProtocol: aws.String("-1"),
						FromPort:   aws.Int32(0),
						ToPort:     aws.Int32(65535),
					},
				},
			}

			obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules)
			awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId

		}()

	case consts.RoleDs:
		func() {

			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules)
			egressrules := ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: SecurityGroup.GroupId,
				IpPermissions: []types.IpPermission{
					{
						IpProtocol: aws.String("-1"),
						FromPort:   aws.Int32(0),
						ToPort:     aws.Int32(65535),
					},
				},
			}

			obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules)
			awsCloudState.InfoDatabase.NetworkSecurityGroup = *SecurityGroup.GroupId
		}()

	case consts.RoleLb:
		func() {

			ingressrules := &ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), *ingressrules)
			if err != nil {
				log.Error("Error creating security group", "error", err)
			}

			egressrules := &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: SecurityGroup.GroupId,
				IpPermissions: []types.IpPermission{
					{
						IpProtocol: aws.String("-1"),
						FromPort:   aws.Int32(0),
						ToPort:     aws.Int32(65535),
					},
				},
			}

			err = obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), *egressrules)
			if err != nil {
				log.Error("Error creating security group", "error", err)
			}

			awsCloudState.InfoLoadBalancer.NetworkSecurityGroup = *SecurityGroup.GroupId
		}()

	default:
		return "", fmt.Errorf("invalid role")

	}

	return *SecurityGroup.GroupId, nil
}
