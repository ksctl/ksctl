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
	} else {

		if len(mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName) == 0 {
			log.Print(awsCtx, "Skipping deleting EKS node-group.")
		} else {
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
		}

		clusterPerimeter := eks.DeleteClusterInput{
			Name: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
		}

		_, err := obj.client.BeginDeleteManagedCluster(awsCtx, &clusterPerimeter)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.ManagedClusterName = ""
		mainStateDocument.CloudInfra.Aws.ManagedClusterArn = ""
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
	}

	iamParameter := iam.DeleteRoleInput{
		RoleName: aws.String(mainStateDocument.CloudInfra.Aws.IamRoleNameWP),
	}

	_, err := obj.client.BeginDeleteIAM(awsCtx, &iamParameter, "worker")
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.IamRoleNameWP = ""
	mainStateDocument.CloudInfra.Aws.IamRoleArnWP = ""
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	iamParameter = iam.DeleteRoleInput{
		RoleName: aws.String(mainStateDocument.CloudInfra.Aws.IamRoleNameCN),
	}

	_, err = obj.client.BeginDeleteIAM(awsCtx, &iamParameter, "controlplane")
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.IamRoleNameCN = ""
	mainStateDocument.CloudInfra.Aws.IamRoleArnCN = ""
	err = storage.Write(mainStateDocument)
	if err != nil {
		return err
	}

	log.Success(awsCtx, "Deleted the EKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	return storage.Write(mainStateDocument)
}

func (obj *AwsProvider) NewManagedCluster(storage types.StorageFactory, noOfNode int) error {
	name := <-obj.chResName
	vmType := <-obj.chVMType

	iamRoleControlPlane := fmt.Sprintf("ksctl-%s-cp-role", name)
	iamRoleWorkerPlane := fmt.Sprintf("ksctl-%s-wp-role", mainStateDocument.CloudInfra.Aws.ManagedClusterName)

	log.Print(awsCtx, "Creating a new EKS cluster.", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)

	if len(mainStateDocument.CloudInfra.Aws.ManagedClusterName) != 0 {
		log.Print(awsCtx, "skipped already created AKS cluster", "name", mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	} else {

		if len(mainStateDocument.CloudInfra.Aws.IamRoleNameCN) == 0 {
			iamParameter := iam.CreateRoleInput{
				RoleName:                 aws.String(iamRoleControlPlane),
				AssumeRolePolicyDocument: aws.String(assumeClusterRolePolicyDocument),
			}
			iamRespCp, err := obj.client.BeginCreateIAM(awsCtx, "controlplane", &iamParameter)
			if err != nil {
				return err
			}

			mainStateDocument.CloudInfra.Aws.IamRoleNameCN = *iamRespCp.Role.RoleName
			mainStateDocument.CloudInfra.Aws.IamRoleArnCN = *iamRespCp.Role.Arn

			err = storage.Write(mainStateDocument)
			if err != nil {
				return err
			}

			log.Success(awsCtx, "created the EKS controlplane role", "name", mainStateDocument.CloudInfra.Aws.IamRoleNameCN)
		} else {
			log.Print(awsCtx, "skipped already created EKS controlplane role", "name", mainStateDocument.CloudInfra.Aws.IamRoleNameCN)
		}

		parameter := eks.CreateClusterInput{
			Name:    aws.String(name),
			RoleArn: aws.String(mainStateDocument.CloudInfra.Aws.IamRoleArnCN),
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

		log.Print(awsCtx, "creating the EKS Controlplane")
		clusterResp, err := obj.client.BeginCreateEKS(awsCtx, &parameter)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.ManagedClusterName = *clusterResp.Cluster.Name
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
	}

	if len(mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName) != 0 {
		log.Print(awsCtx, "skipped already created EKS nodegroup", "name", mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName)
	} else {
		if len(mainStateDocument.CloudInfra.Aws.IamRoleNameWP) == 0 {
			iamParameter := iam.CreateRoleInput{
				RoleName:                 aws.String(iamRoleWorkerPlane),
				AssumeRolePolicyDocument: aws.String(assumeWorkerNodeRolePolicyDocument),
			}
			iamRespWp, err := obj.client.BeginCreateIAM(awsCtx, "worker", &iamParameter)
			if err != nil {
				return err
			}

			mainStateDocument.CloudInfra.Aws.IamRoleNameWP = *iamRespWp.Role.RoleName
			mainStateDocument.CloudInfra.Aws.IamRoleArnWP = *iamRespWp.Role.Arn
			err = storage.Write(mainStateDocument)
			if err != nil {
				return err
			}

			log.Success(awsCtx, "created the EKS worker role", "name", mainStateDocument.CloudInfra.Aws.IamRoleNameWP)
		} else {
			log.Print(awsCtx, "skipped already created ROLE EKS Worker ", "name", mainStateDocument.CloudInfra.Aws.IamRoleNameWP)
		}

		eksNodeGroupName := mainStateDocument.CloudInfra.Aws.ManagedClusterName + "-nodegroup"
		nodegroup := eks.CreateNodegroupInput{
			ClusterName:   aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
			NodeRole:      aws.String(mainStateDocument.CloudInfra.Aws.IamRoleArnWP),
			NodegroupName: aws.String(eksNodeGroupName),
			Subnets:       mainStateDocument.CloudInfra.Aws.SubnetIDs,
			CapacityType:  eks_types.CapacityTypesOnDemand,

			InstanceTypes: []string{vmType},
			// TODO ADD  DISK SIZE OPTION
			DiskSize: aws.Int32(30),

			ScalingConfig: &eks_types.NodegroupScalingConfig{
				DesiredSize: aws.Int32(2),
				MaxSize:     aws.Int32(2), // TODO(praful): need to use the no of Workernodes from the user input
				MinSize:     aws.Int32(2),
			},
		}
		log.Print(awsCtx, "creating the EKS nodegroup")

		nodeResp, err := obj.client.BeginCreateNodeGroup(awsCtx, &nodegroup)
		if err != nil {
			return err
		}

		mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName = *nodeResp.Nodegroup.NodegroupName
		mainStateDocument.CloudInfra.Aws.ManagedNodeGroupArn = *nodeResp.Nodegroup.NodegroupArn

		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
		log.Success(awsCtx, "created the EKS nodegroup", "name", mainStateDocument.CloudInfra.Aws.ManagedNodeGroupName)

	}

	kubeconfig, err := obj.client.GetKubeConfig(awsCtx, mainStateDocument.CloudInfra.Aws.ManagedClusterName)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Aws.B.IsCompleted = true

	mainStateDocument.ClusterKubeConfig = kubeconfig

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	return nil
}
