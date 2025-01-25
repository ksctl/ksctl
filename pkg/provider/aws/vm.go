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
	"encoding/base64"
	"strconv"

	"github.com/ksctl/ksctl/pkg/provider"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
)

func (p *Provider) DelVM(index int) error {

	role := <-p.chRole
	indexNo := index

	p.l.Debug(p.ctx, "Printing", "role", role, "indexNo", indexNo)

	vmName := ""

	switch role {
	case consts.RoleCp:
		vmName = p.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo]
	case consts.RoleDs:
		vmName = p.state.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo]
	case consts.RoleLb:
		vmName = p.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID
	case consts.RoleWp:
		vmName = p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo]
	}

	if len(vmName) == 0 {
		p.l.Success(p.ctx, "skipped already deleted the vm")
	} else {

		var errDel error
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)

			err := p.client.BeginDeleteVM(vmName)
			if err != nil {
				errDel = err
				return
			}
			p.mu.Lock()
			defer p.mu.Unlock()

			switch role {
			case consts.RoleWp:
				p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = ""
				p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[indexNo] = ""
			case consts.RoleCp:
				p.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = ""
				p.state.CloudInfra.Aws.InfoControlPlanes.VMSizes[indexNo] = ""
			case consts.RoleLb:
				p.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID = ""
				p.state.CloudInfra.Aws.InfoLoadBalancer.VMSize = ""
			case consts.RoleDs:
				p.state.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = ""
				p.state.CloudInfra.Aws.InfoDatabase.VMSizes[indexNo] = ""
			}

			err = p.DeleteNetworkInterface(indexNo, role)
			if err != nil {
				errDel = err
				return
			}

			if err := p.store.Write(p.state); err != nil {
				errDel = err
				return
			}
		}()
		<-donePoll
		if errDel != nil {
			return errDel
		}
		p.l.Success(p.ctx, "Deleted the vm", "id", vmName)
	}
	return nil
}

func (p *Provider) CreateNetworkInterface(resName string, index int, role consts.KsctlRole) (string, error) {

	securitygroup, err := p.fetchgroupid(role)
	if err != nil {
		return "", err
	}
	nicid := ""
	switch role {
	case consts.RoleWp:
		nicid = p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		nicid = p.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		nicid = p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId
	case consts.RoleDs:
		nicid = p.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index]
	}

	if len(nicid) != 0 {
		p.l.Print(p.ctx, "skipped already created the network interface", "id", nicid)
		return nicid, nil
	}

	interfaceparameter := &ec2.CreateNetworkInterfaceInput{
		Description: aws.String("network interface"),
		SubnetId:    aws.String(p.state.CloudInfra.Aws.SubnetIDs[0]),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("network-interface"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(string(role) + strconv.Itoa(index) + resName),
					},
				},
			},
		},
		Groups: []string{
			securitygroup,
		},
	}

	nicresponse, err := p.client.BeginCreateNIC(p.ctx, interfaceparameter)
	if err != nil {
		return "", err
	}

	var errCreate error
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleCp:
			p.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleLb:
			p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleDs:
			p.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		}
		if err := p.store.Write(p.state); err != nil {
			errCreate = err
		}
	}()
	<-done
	if errCreate != nil {
		return "", errCreate
	}
	p.l.Success(p.ctx, "Created network interface", "id", *nicresponse.NetworkInterface.NetworkInterfaceId)
	return *nicresponse.NetworkInterface.NetworkInterfaceId, nil
}

func (p *Provider) NewVM(index int) error {
	name := <-p.chResName
	indexNo := index
	role := <-p.chRole
	vmtype := <-p.chVMType

	p.l.Debug(p.ctx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	instanceId := ""
	instanceIp := ""
	switch role {
	case consts.RoleCp:
		instanceId = p.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo]
		instanceIp = p.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo]
	case consts.RoleDs:
		instanceId = p.state.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo]
		instanceIp = p.state.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo]
	case consts.RoleLb:
		instanceId = p.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID
		instanceIp = p.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP
	case consts.RoleWp:
		instanceId = p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo]
		instanceIp = p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo]
	}
	if len(instanceId) != 0 {
		if len(instanceIp) != 0 {
			p.l.Print(p.ctx, "skipped vm already created", "name", instanceId)
			return nil
		} else {
			instanceIp, err := p.client.DescribeInstanceState(p.ctx, instanceId)
			if err != nil {
				return err
			}

			publicip := instanceIp.Reservations[0].Instances[0].PublicIpAddress
			privateip := instanceIp.Reservations[0].Instances[0].PrivateIpAddress

			p.mu.Lock()
			defer p.mu.Unlock()

			switch role {
			case consts.RoleWp:
				p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
				p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
			case consts.RoleCp:
				p.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo] = *publicip
				p.state.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
			case consts.RoleLb:
				p.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP = *publicip
				p.state.CloudInfra.Aws.InfoLoadBalancer.PrivateIP = *privateip
			case consts.RoleDs:
				p.state.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo] = *publicip
				p.state.CloudInfra.Aws.InfoDatabase.PrivateIPs[indexNo] = *privateip
			}
		}
	}

	nicid, err := p.CreateNetworkInterface(name, indexNo, role)
	if err != nil {
		return err
	}

	ami, err := p.getLatestUbuntuAMI()
	if err != nil {
		return err
	}
	initScript, err := provider.CloudInitScript(name)
	if err != nil {
		return err
	}
	initScriptBase64 := base64.StdEncoding.EncodeToString([]byte(initScript))

	parameter := &ec2.RunInstancesInput{
		ImageId:      aws.String(ami),
		InstanceType: types.InstanceType(vmtype),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String(p.state.CloudInfra.Aws.B.SSHKeyName),

		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &types.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					VolumeType:          types.VolumeTypeGp3,
					Throughput:          aws.Int32(125),
					VolumeSize:          aws.Int32(30),
					Iops:                aws.Int32(3000),
				},
			},
		},

		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("instance"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(name),
					},
				},
			},
		},

		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:        aws.Int32(0),
				NetworkInterfaceId: aws.String(nicid),
			},
		},
		UserData: aws.String(initScriptBase64),
	}

	instanceop, err := p.client.BeginCreateVM(p.ctx, parameter)
	if err != nil {
		return err
	}

	instanceId = *instanceop.Instances[0].InstanceId

	var errCreateVM error

	done1 := make(chan struct{})
	go func() {
		defer close(done1)
		p.mu.Lock()
		defer p.mu.Unlock()
		switch role {
		case consts.RoleCp:
			p.state.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = instanceId
			p.state.CloudInfra.Aws.InfoControlPlanes.HostNames[indexNo] = name
			p.state.CloudInfra.Aws.InfoControlPlanes.VMSizes[indexNo] = vmtype
		case consts.RoleDs:
			p.state.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = instanceId
			p.state.CloudInfra.Aws.InfoDatabase.HostNames[indexNo] = name
			p.state.CloudInfra.Aws.InfoDatabase.VMSizes[indexNo] = vmtype
		case consts.RoleLb:
			p.state.CloudInfra.Aws.InfoLoadBalancer.InstanceID = instanceId
			p.state.CloudInfra.Aws.InfoLoadBalancer.HostName = name
			p.state.CloudInfra.Aws.InfoLoadBalancer.VMSize = vmtype
		case consts.RoleWp:
			p.state.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = instanceId
			p.state.CloudInfra.Aws.InfoWorkerPlanes.HostNames[indexNo] = name
			p.state.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[indexNo] = vmtype
		}

		if err := p.store.Write(p.state); err != nil {
			errCreateVM = err
		}
	}()
	<-done1
	if errCreateVM != nil {
		return errCreateVM
	}

	p.l.Print(p.ctx, "creating vm", "vmName", name)

	done := make(chan struct{})

	go func() {
		defer close(done)

		err = p.client.InstanceInitialWaiter(p.ctx, instanceId)
		if err != nil {
			errCreateVM = err
			return
		}

		instanceIp, err := p.client.DescribeInstanceState(p.ctx, instanceId)
		if err != nil {
			errCreateVM = err
			return
		}

		publicip := instanceIp.Reservations[0].Instances[0].PublicIpAddress
		privateip := instanceIp.Reservations[0].Instances[0].PrivateIpAddress

		p.mu.Lock()
		defer p.mu.Unlock()

		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
			p.state.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleCp:
			p.state.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo] = *publicip
			p.state.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleLb:
			p.state.CloudInfra.Aws.InfoLoadBalancer.PublicIP = *publicip
			p.state.CloudInfra.Aws.InfoLoadBalancer.PrivateIP = *privateip
		case consts.RoleDs:
			p.state.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo] = *publicip
			p.state.CloudInfra.Aws.InfoDatabase.PrivateIPs[indexNo] = *privateip
		}
		if err := p.store.Write(p.state); err != nil {
			errCreateVM = err
			return
		}
	}()
	<-done
	if errCreateVM != nil {
		return errCreateVM
	}

	p.l.Success(p.ctx, "Created the vm", "name", name)
	return nil
}

func (p *Provider) getLatestUbuntuAMI() (string, error) {
	imageFilter := &ec2.DescribeImagesInput{ // https://cloud-images.ubuntu.com/locator/ec2/
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
			{
				Name:   aws.String("owner-alias"),
				Values: []string{"amazon"},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
			{
				Name:   aws.String("is-public"),
				Values: []string{"true"},
			},
		},
	}

	resp, err := p.client.FetchLatestAMIWithFilter(imageFilter)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (p *Provider) DeleteNetworkInterface(index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		interfaceName = p.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		interfaceName = p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId
	case consts.RoleDs:
		interfaceName = p.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index]
	}
	if len(interfaceName) == 0 {
		p.l.Print(p.ctx, "skipped already deleted the network interface")
	} else {
		err := p.client.BeginDeleteNIC(interfaceName)
		if err != nil {
			return err
		}
		switch role {
		case consts.RoleWp:
			p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
		case consts.RoleCp:
			p.state.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
		case consts.RoleLb:
			p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId = ""
		case consts.RoleDs:
			p.state.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index] = ""
		}
		err = p.store.Write(p.state)
		if err != nil {
			return err
		}
		p.l.Success(p.ctx, "deleted the network interface", "id", interfaceName)
	}

	return nil
}

func (p *Provider) fetchgroupid(role consts.KsctlRole) (string, error) {
	switch role {
	case consts.RoleCp:
		return p.state.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs, nil
	case consts.RoleWp:
		return p.state.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs, nil
	case consts.RoleLb:
		return p.state.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID, nil
	case consts.RoleDs:
		return p.state.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs, nil

	}

	return "", ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidKsctlRole,
		p.l.NewError(p.ctx, "invalid role", "role", role),
	)
}
