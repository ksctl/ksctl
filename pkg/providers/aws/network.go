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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func (p *Provider) DelNetwork() error {

	if len(p.state.CloudInfra.Aws.SubnetIDs) == 0 {
		p.l.Print(p.ctx, "skipped already deleted the vpc", "name", p.state.CloudInfra.Aws.VpcName)
	} else {
		err := p.DeleteSubnet(p.state.CloudInfra.Aws.SubnetIDs)
		if err != nil {
			return err
		}
	}

	err := p.client.BeginDeleteVirtNet(p.ctx)
	if err != nil {
		return err
	}

	if p.state.CloudInfra.Aws.VpcId == "" {
		p.l.Success(p.ctx, "Deleted the vpc", "id", p.state.CloudInfra.Aws.VpcName)
	} else {
		err = p.DeleteVpc()
		if err != nil {
			return err
		}
	}

	p.l.Success(p.ctx, "Deleted the vpc", "name", p.state.CloudInfra.Aws.VpcName)

	if err := p.store.DeleteCluster(); err != nil {
		return err
	}

	return nil
}

func (p *Provider) DeleteSubnet(subnetID []string) error {

	for i := 0; i < len(p.state.CloudInfra.Aws.SubnetIDs); i++ {
		err := p.client.BeginDeleteSubNet(p.ctx, subnetID[i])
		if err != nil {
			return err
		}
		p.state.CloudInfra.Aws.SubnetIDs[i] = ""

		if err := p.store.Write(p.state); err != nil {
			return err
		}
	}

	p.l.Success(p.ctx, "Deleted the subnet", "id", p.state.CloudInfra.Aws.SubnetNames)

	return nil
}

func (p *Provider) DeleteVpc() error {

	err := p.client.BeginDeleteVpc(p.ctx)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Aws.VpcId = ""
	name := p.state.CloudInfra.Aws.VpcName
	p.state.CloudInfra.Aws.VpcName = ""
	if err := p.store.Write(p.state); err != nil {
		return err
	}

	p.l.Success(p.ctx, "Deleted the vpc", "name", name)
	return nil
}

func (p *Provider) NewNetwork() error {
	<-p.chResName

	if len(p.state.CloudInfra.Aws.VpcId) != 0 {
		p.l.Print(p.ctx, "skipped already created the vpc", p.state.CloudInfra.Aws.VpcName)
	} else {
		vpcclient := ec2.CreateVpcInput{
			CidrBlock: aws.String("172.31.0.0/16"),
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceType("vpc"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(p.ClusterName + "-vpc"),
						},
					},
				},
			},
		}
		p.state.CloudInfra.Aws.VpcCidr = "172.31.0.0/16"

		p.l.Debug(p.ctx, "Printing", "virtualprivatecloud", vpcclient)

		vpc, err := p.client.BeginCreateVpc(vpcclient)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Aws.VpcId = *vpc.Vpc.VpcId
		p.state.CloudInfra.Aws.VpcName = *vpc.Vpc.Tags[0].Value

		if err := p.client.ModifyVpcAttribute(p.ctx); err != nil {
			return err
		}

		if err := p.store.Write(p.state); err != nil {
			return err
		}

		p.l.Success(p.ctx, "created the vpc", "id", *vpc.Vpc.VpcId)

	}

	subNet := p.ClusterName + "-subnet"

	if err := p.CreateSubnet(subNet); err != nil {
		return err
	}

	if err := p.CreateVirtualNetwork(); err != nil {
		return err
	}

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	return nil
}

func (p *Provider) CreateSubnet(subnetName string) error {

	zones, err := p.client.GetAvailabilityZones()
	if err != nil {
		return err
	}

	subnets := []string{"172.31.0.0/20", "172.31.32.0/20", "172.31.16.0/20"}

	if len(p.state.CloudInfra.Aws.SubnetIDs) != 0 {
		p.l.Print(p.ctx, "skipped already created the subnet", p.state.CloudInfra.Aws.SubnetIDs)
	} else {
		for i := 0; i < 3; i++ {
			parameter := ec2.CreateSubnetInput{
				CidrBlock: aws.String(subnets[i]),
				VpcId:     aws.String(p.state.CloudInfra.Aws.VpcId),

				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceType("subnet"),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(p.ClusterName + "-subnet" + strconv.Itoa(i)),
							},
						},
					},
				},
				AvailabilityZone: aws.String(*zones.AvailabilityZones[i].ZoneName),
			}

			p.l.Print(p.ctx, "Selected availability zone", "zone", *zones.AvailabilityZones[i].ZoneName)
			response, err := p.client.BeginCreateSubNet(p.ctx, subnetName, parameter)
			if err != nil {
				return err
			}

			p.state.CloudInfra.Aws.SubnetIDs = append(p.state.CloudInfra.Aws.SubnetIDs, *response.Subnet.SubnetId)
			p.state.CloudInfra.Aws.SubnetNames = append(p.state.CloudInfra.Aws.SubnetNames, *response.Subnet.Tags[0].Value)

			if err := p.client.ModifySubnetAttribute(p.ctx, i); err != nil {
				return err
			}

			if err := p.store.Write(p.state); err != nil {
				return err
			}

			p.l.Success(p.ctx, "created the subnet", "id", *response.Subnet.Tags[0].Value)
		}

		naclinput := ec2.CreateNetworkAclInput{
			VpcId: aws.String(p.state.CloudInfra.Aws.VpcId),
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceType("network-acl"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(p.ClusterName + "-nacl"),
						},
					},
				},
			},
		}

		naclresp, err := p.client.BeginCreateNetworkAcl(p.ctx, naclinput)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Aws.NetworkAclID = *naclresp.NetworkAcl.NetworkAclId

		if err := p.store.Write(p.state); err != nil {
			return err
		}

		p.l.Success(p.ctx, "created the network acl", "id", *naclresp.NetworkAcl.NetworkAclId)
	}

	return nil
}

func (p *Provider) CreateVirtualNetwork() error {

	if len(p.state.CloudInfra.Aws.RouteTableID) != 0 {
		p.l.Success(p.ctx, "skipped already created the route table ", "id", p.state.CloudInfra.Aws.RouteTableID)
	} else {
		internetGateway := ec2.CreateInternetGatewayInput{
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceType("internet-gateway"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(p.ClusterName + "-ig"),
						},
					},
				},
			},
		}

		routeTableClient := ec2.CreateRouteTableInput{
			VpcId: aws.String(p.state.CloudInfra.Aws.VpcId),
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceType("route-table"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(p.ClusterName + "-rt"),
						},
					},
				},
			},
		}

		routeresponce, gatewayresp, err := p.client.BeginCreateVirtNet(internetGateway, routeTableClient, p.state.CloudInfra.Aws.VpcId)
		if err != nil {
			return err
		}

		p.state.CloudInfra.Aws.RouteTableID = *routeresponce.RouteTable.RouteTableId
		p.state.CloudInfra.Aws.GatewayID = *gatewayresp.InternetGateway.InternetGatewayId

		if err := p.store.Write(p.state); err != nil {
			return err
		}
		p.l.Success(p.ctx, "created the internet gateway", "id", *gatewayresp.InternetGateway.InternetGatewayId)
		p.l.Success(p.ctx, "created the route table", "id", *routeresponce.RouteTable.RouteTableId)
	}

	return nil
}
