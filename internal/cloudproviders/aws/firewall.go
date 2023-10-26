package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func (obj *AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	// name := obj.metadata.resName
	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	switch role {
	case RoleCp:
		GroupID, err := obj.CreateSecurityGroup(obj.metadata.role)
		if err != nil {
			fmt.Println(err)
		}

		firewallRuleControlPlane(obj.ec2Client(), GroupID)

	case RoleWp:
		GroupID, err := obj.CreateSecurityGroup(role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleWorkerPlane(obj.ec2Client(), GroupID)

	case RoleLb:
		Groupid, err := obj.CreateSecurityGroup(role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleLoadBalancer(obj.ec2Client(), Groupid)
	case RoleDs:
		GroupID, err := obj.CreateSecurityGroup(role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleDataStore(obj.ec2Client(), GroupID)

	default:
		return fmt.Errorf("invalid role")
	}

	return nil
}

func (obj *AwsProvider) CreateSecurityGroup(Role KsctlRole) (string, error) {
	SecurityGroup, err := obj.ec2Client().CreateSecurityGroup(context.Background(), &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(string(Role + "securitygroup")),
		Description: aws.String(obj.clusterName + "securitygroup"),
		VpcId:       aws.String(awsCloudState.VPCID),
	})
	if err != nil {
		fmt.Println(err)
	}

	switch Role {
	case RoleCp:
		awsCloudState.SecurityGroupID[0] = *SecurityGroup.GroupId
		awsCloudState.SecurityGroupRole[0] = string(Role)
	case RoleWp:
		awsCloudState.SecurityGroupID[1] = *SecurityGroup.GroupId
		awsCloudState.SecurityGroupRole[1] = string(Role)
	case RoleDs:
		awsCloudState.SecurityGroupID[2] = *SecurityGroup.GroupId
		awsCloudState.SecurityGroupRole[2] = string(Role)
	case RoleLb:
		awsCloudState.SecurityGroupID[3] = *SecurityGroup.GroupId
		awsCloudState.SecurityGroupRole[3] = string(Role)
	default:
		return "", fmt.Errorf("invalid role")

	}

	return *SecurityGroup.GroupId, nil
}

// TODO ADD FIREWALL RULES
func firewallRuleControlPlane(client *ec2.Client, GroupId string) {
	// inbound rules
	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

	// outbound rules
	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

}
func firewallRuleWorkerPlane(client *ec2.Client, GroupId string) {
	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

	// outbound rules
	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

}

func firewallRuleDataStore(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(6443),
				ToPort:     aws.Int32(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(2379),
				ToPort:     aws.Int32(2380),
			},
		},
	})

}

func firewallRuleLoadBalancer(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(0),
		ToPort:     aws.Int32(65535),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(0),
		ToPort:     aws.Int32(65535),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

}
