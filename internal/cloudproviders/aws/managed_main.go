package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eks_types "github.com/aws/aws-sdk-go-v2/service/eks/types"
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

	parameter := eks.CreateClusterInput{
		Name: aws.String(name),
		ResourcesVpcConfig: &eks_types.VpcConfigRequest{
			EndpointPrivateAccess: aws.Bool(true),
			EndpointPublicAccess:  aws.Bool(true),
			PublicAccessCidrs:     []string{"0.0.0.0/0"},
		},
		KubernetesNetworkConfig: &eks_types.KubernetesNetworkConfigRequest{
			IpFamily:        eks_types.IpFamilyIpv4,
			ServiceIpv4Cidr: aws.String("0.0.0.0/0"),
		},
		Version: aws.String(""),
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
		ClusterName: aws.String(mainStateDocument.CloudInfra.Aws.ManagedClusterName),
		//NodeRole:      aws.String(*result.Role.RoleId),
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
