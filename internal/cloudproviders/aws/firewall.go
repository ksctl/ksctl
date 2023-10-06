package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func (obj AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	switch obj.metadata.role {
	case utils.ROLE_CP:
		GroupID, err := obj.CreateSecurityGroup(obj.metadata.role)
		if err != nil {
			fmt.Println(err)
		}

		firewallRuleControlPlane(obj.ec2Client(), GroupID)
		awsCloudState.SecurityGroupID = GroupID

	case utils.ROLE_WP:
		firewallRuleWorkerPlane()

	case utils.ROLE_LB:
		Groupid, err := obj.CreateSecurityGroup(obj.metadata.role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleLoadBalancer(obj.ec2Client(), Groupid)
	case utils.ROLE_DS:
		firewallRuleDataStore()

	case "testing":
		GroupID, err := obj.CreateSecurityGroup(obj.metadata.role)
		if err != nil {
			fmt.Println(err)
		}
		TESTING(obj.ec2Client(), GroupID)

	default:
		return fmt.Errorf("invalid role")
	}

	return nil
}

func TESTING(client *ec2.Client, GroupId string) {
	_, err := client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: &GroupId,
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (obj AwsProvider) CreateSecurityGroup(Role string) (string, error) {
	SecurityGroup, err := obj.ec2Client().CreateSecurityGroup(context.Background(), &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(obj.clusterName + Role + "securitygroup"),
		Description: aws.String(obj.clusterName + "securitygroup"),
		VpcId:       aws.String(awsCloudState.VPCID),
	})
	if err != nil {
		fmt.Println(err)
	}

	awsCloudState.SecurityGroupName = (obj.clusterName + obj.metadata.role + "securitygroup")
	awsCloudState.SecurityGroupID = *SecurityGroup.GroupId

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
func firewallRuleWorkerPlane() {
	// creating firewall rule for workerplane

}

func firewallRuleDataStore() {
	// creating firewall rule for dataservices

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
