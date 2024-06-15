package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eks_types "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/ksctl/ksctl/pkg/types"
)

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
	mainStateDocument.CloudInfra.Aws.ManagedNodeGroupArn = ""
	mainStateDocument.CloudInfra.Aws.NoManagedNodes = 0
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	clusterPerimeter := eks.DeleteClusterInput{
		Name: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
	}

	_, err = obj.client.BeginDeleteManagedCluster(awsCtx, &clusterPerimeter)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.ManagedClusterName = ""
	mainStateDocument.CloudInfra.Aws.ManagedClusterArn = ""
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	iamParameter := iam.DeleteRoleInput{
		RoleName: aws.String(mainStateDocument.CloudInfra.Aws.IamRoleName),
	}

	_, err = obj.client.BeginDeleteIAM(awsCtx, &iamParameter)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.IamRoleName = ""
	mainStateDocument.CloudInfra.Aws.IamRoleArn = ""
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the EKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	return storage.Write(mainStateDocument)
}

func (obj *AwsProvider) NewManagedCluster(storage types.StorageFactory, noOfNode int) error {
	name := <-obj.chResName
	vmtype := <-obj.chVMType

	log.Debug(awsCtx, "Creating a new EKS cluster.", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)

	if len(mainStateDocument.CloudInfra.Aws.ManagedClusterName) != 0 {
		log.Print(awsCtx, "skipped already created AKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	} else {
		iamParameter := iam.CreateRoleInput{
			RoleName:                 aws.String(mainStateDocument.ClusterName + "-controlplane" + "role"),
			AssumeRolePolicyDocument: aws.String(assumeClusterRolePolicyDocument),
		}
		iamRespCp, err := obj.client.BeginCreateIAM(awsCtx, "controlplane", &iamParameter)
		if err != nil {
			return err
		}

		parameter := eks.CreateClusterInput{
			Name:    aws.String(name),
			RoleArn: aws.String(*iamRespCp.Role.Arn),
			ResourcesVpcConfig: &eks_types.VpcConfigRequest{
				EndpointPrivateAccess: aws.Bool(true),
				EndpointPublicAccess:  aws.Bool(true),
				PublicAccessCidrs:     []string{"0.0.0.0/0"},
				SubnetIds:             mainStateDocument.CloudInfra.Aws.SubnetIDs,
			},
			KubernetesNetworkConfig: &eks_types.KubernetesNetworkConfigRequest{
				IpFamily: eks_types.IpFamilyIpv4,
			},
			AccessConfig: &eks_types.CreateAccessConfigRequest{
				AuthenticationMode:                      eks_types.AuthenticationModeApi,
				BootstrapClusterCreatorAdminPermissions: aws.Bool(true),
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
	}

	if len(mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName) != 0 {
		log.Print(awsCtx, "skipped already created AKS node-group", "name", mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName)
	} else {
		iamParameter := iam.CreateRoleInput{
			RoleName:                 aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName + "-worker" + "role"),
			AssumeRolePolicyDocument: aws.String(assumeWorkerNodeRolePolicyDocument),
		}
		iamRespWp, err := obj.client.BeginCreateIAM(awsCtx, "worker", &iamParameter)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.IamRoleName = *iamRespWp.Role.RoleName
		mainStateDocument.CloudInfra.Aws.IamRoleArn = *iamRespWp.Role.Arn
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}

		nodegroup := eks.CreateNodegroupInput{
			ClusterName:   aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
			NodeRole:      aws.String(*iamRespWp.Role.Arn),
			NodegroupName: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName + "-nodegroup"),
			Subnets:       mainStateDocument.CloudInfra.Aws.SubnetIDs,
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
		mainStateDocument.CloudInfra.Aws.ManagedNodeGroupArn = *nodeResp.Nodegroup.NodegroupArn

		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
	}

	result, err := obj.client.DescribeCluster(awsCtx, &eks.DescribeClusterInput{
		Name: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
	})
	if err != nil {
		return err
	}

	config := result.Cluster.AccessConfig

	// create get kubeconfig funtion
	// kubeconfig, err := obj.client.GetKubeConfig(awsCtx, mainStateDocument.CloudInfra.Aws.ManagedClusterName)

	return nil
}
