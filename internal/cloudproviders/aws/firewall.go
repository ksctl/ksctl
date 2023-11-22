package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func (obj *AwsProvider) NewFirewall(storage resources.StorageFactory) error {

	// name := obj.metadata.resName
	role := obj.metadata.role
	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	switch role {
	case consts.RoleCp:
		GroupID, err := obj.CreateSecurityGroup(obj.metadata.role)
		if err != nil {
			fmt.Println(err)
		}

		firewallRuleControlPlane(obj.ec2Client(), GroupID)

	case consts.RoleWp:
		GroupID, err := obj.CreateSecurityGroup(role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleWorkerPlane(obj.ec2Client(), GroupID)

	case consts.RoleLb:
		Groupid, err := obj.CreateSecurityGroup(role)
		if err != nil {
			fmt.Println(err)
		}
		firewallRuleLoadBalancer(obj.ec2Client(), Groupid)
	case consts.RoleDs:
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

		log.Success("[aws] deleted the security group ", nsg)
	}

	return nil
}

func (obj *AwsProvider) CreateSecurityGroup(Role consts.KsctlRole) (string, error) {
	SecurityGroup, err := obj.ec2Client().CreateSecurityGroup(context.Background(), &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(string(Role + "securitygroup")),
		Description: aws.String(obj.clusterName + "securitygroup"),
		VpcId:       aws.String(awsCloudState.VPCID),
	})
	if err != nil {
		fmt.Println(err)
	}

	switch Role {
	case consts.RoleCp:
		awsCloudState.InfoControlPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleWp:
		awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleDs:
		awsCloudState.InfoDatabase.NetworkSecurityGroup = *SecurityGroup.GroupId
	case consts.RoleLb:
		awsCloudState.InfoLoadBalancer.NetworkSecurityGroup = *SecurityGroup.GroupId
	default:
		return "", fmt.Errorf("invalid role")

	}

	return *SecurityGroup.GroupId, nil
}

// TODO ADD FIREWALL RULES
func firewallRuleControlPlane(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

}
func firewallRuleWorkerPlane(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
}

func firewallRuleDataStore(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
}

func firewallRuleLoadBalancer(client *ec2.Client, GroupId string) {

	client.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	client.AuthorizeSecurityGroupEgress(context.Background(), &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:    &GroupId,
		IpProtocol: aws.String("-1"),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

}