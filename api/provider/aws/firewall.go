package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func (obj AwsProvider) NewFirewall(storage resources.StorageFactory) error {
	ex := ""
	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		ex = "controlplane"
	case utils.ROLE_WP:
		ex = "workerplane"
	case utils.ROLE_LB:
		ex = "loadbalancer"
	case utils.ROLE_DS:
		ex = "dataservices"
	default:
		return fmt.Errorf("invalid role")
	}

	switch obj.Metadata.Role {
	case utils.ROLE_CP:
		GroupID, err := obj.CreateSecurityGroup(obj.Metadata.Role)
		if err != nil {
			fmt.Println(err)
		}

		firewallRuleControlPlane(obj.ec2Client(), GroupID)
		awsCloudState.SecurityGroupID = GroupID

	case utils.ROLE_WP:
		firewallRuleWorkerPlane()

	case utils.ROLE_LB:
		Groupid, err := obj.CreateSecurityGroup(obj.Metadata.Role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleLoadBalancer(obj.ec2Client(), Groupid)

	case utils.ROLE_DS:
		firewallRuleDataStore()
	default:
		return fmt.Errorf("invalid role")
	}

	return nil
}

func (obj AwsProvider) CreateSecurityGroup(Role string) (string, error) {
	SecurityGroup, err := obj.ec2Client().CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(obj.ClusterName + Role + "securitygroup"),
		Description: aws.String(obj.ClusterName + "securitygroup"),
		VpcId:       aws.String(obj.VPC),
	})
	if err != nil {
		fmt.Println(err)
	}

	awsCloudState.SecurityGroupName = (obj.ClusterName + obj.Metadata.Role + "securitygroup")
	awsCloudState.SecurityGroupID = *SecurityGroup.GroupId

	return *SecurityGroup.GroupId, nil
}

// TODO ADD FIREWALL RULES
func firewallRuleControlPlane(client *ec2.EC2, GroupId string) {
	// inbound rules
	client.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(6443),
				ToPort:     aws.Int64(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(2379),
				ToPort:     aws.Int64(2380),
			},
		},
	})

	// outbound rules
	client.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		CidrIp:     aws.String("0.0.0.0/0"),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(6443),
				ToPort:     aws.Int64(6443),
			}, {
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(2379),
				ToPort:     aws.Int64(2380),
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

func firewallRuleLoadBalancer(client *ec2.EC2, GroupId string) {

	client.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(0),
		ToPort:     aws.Int64(65535),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(0),
		ToPort:     aws.Int64(65535),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

}
