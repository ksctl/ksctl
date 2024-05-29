package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eks_types "github.com/aws/aws-sdk-go-v2/service/eks/types"
	iam2 "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/ksctl/ksctl/pkg/types"
)

const eksNodeGroupPolicyDocument = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"eks:CreateNodegroup",
				"eks:DescribeNodegroup",
				"eks:DeleteNodegroup"
			],
			"Resource": "*"
		}
	]
}`

const eksClusterPolicyDocument = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"eks:CreateCluster",
				"eks:DescribeCluster",
				"eks:DeleteCluster"
			],
			"Resource": "*"
		}
	]
}`

const assumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "eks.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`

func (obj *AwsProvider) DelManagedCluster(storage types.StorageFactory) error {
	if len(mainStateDocument.CloudInfra.Aws.ManagedClusterName) == 0 {
		log.Print(awsCtx, "Skipping deleting EKS cluster.")
		return nil
	}

	nodeParameter := eks.DeleteNodegroupInput{
		ClusterName:   aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
		NodegroupName: aws.String(mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName),
	}

	_, err := obj.client.BeginDeleteNodeGroup(awsCtx, &nodeParameter)
	if err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the EKS node-group", "name", mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName)

	mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName = ""
	mainStateDocument.CloudInfra.Aws.NoManagedNodes = 0

	clusterPerimeter := eks.DeleteClusterInput{
		Name: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
	}

	_, err = obj.client.BeginDeleteManagedCluster(awsCtx, &clusterPerimeter)
	if err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the EKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	mainStateDocument.CloudInfra.Aws.ManagedClusterName = ""
	return storage.Write(mainStateDocument)
}

func (obj *AwsProvider) NewManagedCluster(storage types.StorageFactory, noOfNode int) error {
	name := <-obj.chResName
	vmtype := <-obj.chVMType

	log.Debug(awsCtx, "Creating a new EKS cluster.", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)

	if len(mainStateDocument.CloudInfra.Aws.ManagedClusterName) != 0 {
		log.Print(awsCtx, "skipped already created AKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	}

	mainStateDocument.CloudInfra.Aws.NoManagedNodes = noOfNode
	mainStateDocument.CloudInfra.Aws.B.KubernetesVer = obj.metadata.k8sVersion
	mainStateDocument.BootstrapProvider = "managed"

	iamParameter := iam2.CreateRoleInput{
		RoleName:                 aws.String(obj.clusterName + "Node-Group"),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDocument),
	}
	iamResp, err := obj.client.BeginCreateIAM(awsCtx, &iamParameter)
	if err != nil {
		return err
	}

	parameter := eks.CreateClusterInput{
		Name:    aws.String(name),
		RoleArn: aws.String(*iamResp.Role.Arn),
		ResourcesVpcConfig: &eks_types.VpcConfigRequest{
			EndpointPrivateAccess: aws.Bool(true),
			EndpointPublicAccess:  aws.Bool(true),
			PublicAccessCidrs:     []string{"0.0.0.0/0"},
		},
		KubernetesNetworkConfig: &eks_types.KubernetesNetworkConfigRequest{
			IpFamily:        eks_types.IpFamilyIpv4,
			ServiceIpv4Cidr: aws.String("0.0.0.0/0"),
		},
		Version: aws.String(obj.metadata.k8sVersion),
	}

	clusterResp, err := obj.client.BeginCreateEKS(awsCtx, &parameter)
	if err != nil {
		return err
	}

	fmt.Println(clusterResp)

	mainStateDocument.CloudInfra.Aws.ManagedClusterName = *clusterResp.Cluster.Name
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	nodegroup := eks.CreateNodegroupInput{
		ClusterName:   aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
		NodeRole:      aws.String(*iamResp.Role.RoleName),
		NodegroupName: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName + "nod-group"),
		Subnets:       []string{mainStateDocument.CloudInfra.Aws.SubnetID},
		CapacityType:  eks_types.CapacityTypesOnDemand,

		InstanceTypes: []string{vmtype},
		// TODO ADD  DISK SIZE OPTION
		DiskSize: aws.Int32(30),

		ScalingConfig: &eks_types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(2),
			MaxSize:     aws.Int32(2),
			MinSize:     aws.Int32(2),
		},
	}

	nodeResp, err := obj.client.BeignCreateNodeGroup(awsCtx, &nodegroup)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName = *nodeResp.Nodegroup.NodegroupName
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}
	fmt.Println(nodeResp)
	return nil
}
