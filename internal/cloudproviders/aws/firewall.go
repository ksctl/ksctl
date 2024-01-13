package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func (obj *AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	role := <-obj.chRole
	_ = <-obj.chResName
	_, err := obj.CreateSecurityGroup(role)
	if err != nil {
		fmt.Println(err)
	}

	if err := storage.Write(mainStateDocument); err != nil {
		log.Error("Error saving state", "error", err)
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
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup = ""
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup = ""
		default:
			return fmt.Errorf("invalid role")
		}

		err = storage.Write(mainStateDocument)
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
		VpcId:       aws.String(mainStateDocument.CloudInfra.Aws.VPCID),
	}

	switch Role {
	case consts.RoleCp:
		if mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleWp:
		if mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup, nil
		}
	case consts.RoleLb:
		if mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup, nil
		}
	case consts.RoleDs:
		if mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup != "" {
			log.Success("[skip] already created the security group ", "id: ", mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup)
			return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup, nil
		}
	}

	SecurityGroup, err := obj.client.BeginCreateSecurityGroup(context.Background(), obj.ec2Client(), SecurityGroupInput)
	if err != nil {
		fmt.Println(err)
	}

	switch Role {
	case consts.RoleCp:
		mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId

		func() {
			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			if err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}

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

			if err := obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}
		}()

	case consts.RoleWp:
		func() {

			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			if err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}
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

			if err := obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId

		}()

	case consts.RoleDs:
		func() {

			ingressrules := ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:    SecurityGroup.GroupId,
				IpProtocol: aws.String("-1"),
				CidrIp:     aws.String("0.0.0.0/0"),
			}

			if err := obj.client.AuthorizeSecurityGroupIngress(context.Background(), obj.ec2Client(), ingressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}
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

			if err := obj.client.AuthorizeSecurityGroupEgress(context.Background(), obj.ec2Client(), egressrules); err != nil {
				log.Error("Error creating security group", "error", err)
			}
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup = *SecurityGroup.GroupId
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

			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup = *SecurityGroup.GroupId
		}()

	default:
		return "", fmt.Errorf("invalid role")

	}

	return *SecurityGroup.GroupId, nil
}
