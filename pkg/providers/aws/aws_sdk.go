// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type CloudSDK interface {
	InitClient(b *Provider) error

	AuthorizeSecurityGroupIngress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupIngressInput) error

	ListLocations() ([]string, error)

	ListVMTypes() (ec2.DescribeInstanceTypesOutput, error)

	BeginCreateVpc(parameter ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error)

	BeginCreateVirtNet(gatewayparameter ec2.CreateInternetGatewayInput, routeTableparameter ec2.CreateRouteTableInput, vpcid string) (*ec2.CreateRouteTableOutput, *ec2.CreateInternetGatewayOutput, error)

	BeginCreateSubNet(context context.Context, subnetName string, parameter ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error)

	BeginDeleteVirtNet(ctx context.Context) error

	BeginDeleteSubNet(ctx context.Context, subnetID string) error

	DeleteSSHKey(ctx context.Context, name string) error

	BeginCreateVM(ctx context.Context, parameter *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)

	BeginDeleteVM(vmname string) error

	BeginCreateNIC(ctx context.Context, parameter *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)

	BeginDeleteNIC(nicID string) error

	BeginDeleteVpc(ctx context.Context) error

	BeginCreateNetworkAcl(ctx context.Context, parameter ec2.CreateNetworkAclInput) (*ec2.CreateNetworkAclOutput, error)

	BeginCreateSecurityGroup(ctx context.Context, parameter ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	BeginDeleteSecurityGrp(ctx context.Context, securityGrpID string) error

	DescribeInstanceState(ctx context.Context, instanceId string) (*ec2.DescribeInstancesOutput, error)

	FetchLatestAMIWithFilter(filter *ec2.DescribeImagesInput) (string, error)
	GetAvailabilityZones() (*ec2.DescribeAvailabilityZonesOutput, error)
	AuthorizeSecurityGroupEgress(ctx context.Context, parameter ec2.AuthorizeSecurityGroupEgressInput) error

	ImportKeyPair(ctx context.Context, keypair *ec2.ImportKeyPairInput) error
	InstanceInitialWaiter(ctx context.Context, instanceInput string) error

	ModifyVpcAttribute(ctx context.Context) error
	ModifySubnetAttribute(ctx context.Context, i int) error

	BeginCreateEKS(ctx context.Context, parameter *eks.CreateClusterInput) (*eks.CreateClusterOutput, error)
	BeginCreateNodeGroup(ctx context.Context, paramter *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error)

	BeginDeleteNodeGroup(ctx context.Context, parameter *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error)
	BeginDeleteManagedCluster(ctx context.Context, parameter *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error)
	DescribeCluster(ctx context.Context, parameter *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error)

	BeginCreateIAM(ctx context.Context, node string, parameter *iam.CreateRoleInput) (*iam.CreateRoleOutput, error)
	BeginDeleteIAM(ctx context.Context, parameter *iam.DeleteRoleInput, node string) (*iam.DeleteRoleOutput, error)

	GetKubeConfig(ctx context.Context, clusterName string) (string, error)

	ListK8sVersions(ctx context.Context) ([]string, error)
}
