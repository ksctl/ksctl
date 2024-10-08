package aws

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	ksctlTypes "github.com/ksctl/ksctl/pkg/types"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
)

func (obj *AwsProvider) DelVM(storage ksctlTypes.StorageFactory, index int) error {

	role := <-obj.chRole
	indexNo := index

	log.Debug(awsCtx, "Printing", "role", role, "indexNo", indexNo)

	vmName := ""

	switch role {
	case consts.RoleCp:
		vmName = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo]
	case consts.RoleDs:
		vmName = mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo]
	case consts.RoleLb:
		vmName = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID
	case consts.RoleWp:
		vmName = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo]
	}

	if len(vmName) == 0 {
		log.Success(awsCtx, "skipped already deleted the vm")
	} else {

		var errDel error
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)

			err := obj.client.BeginDeleteVM(vmName)
			if err != nil {
				errDel = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()

			switch role {
			case consts.RoleWp:
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = ""
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[indexNo] = ""
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = ""
				mainStateDocument.CloudInfra.Aws.InfoControlPlanes.VMSizes[indexNo] = ""
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID = ""
				mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.VMSize = ""
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = ""
				mainStateDocument.CloudInfra.Aws.InfoDatabase.VMSizes[indexNo] = ""
			}

			err = obj.DeleteNetworkInterface(awsCtx, storage, indexNo, role)
			if err != nil {
				errDel = err
				return
			}

			if err := storage.Write(mainStateDocument); err != nil {
				errDel = err
				return
			}
		}()
		<-donePoll
		if errDel != nil {
			return errDel
		}
		log.Success(awsCtx, "Deleted the vm", "id", vmName)
	}
	return nil
}

func (obj *AwsProvider) CreateNetworkInterface(ctx context.Context, storage ksctlTypes.StorageFactory, resName string, index int, role consts.KsctlRole) (string, error) {

	securitygroup, err := fetchgroupid(role)
	if err != nil {
		return "", err
	}
	nicid := ""
	switch role {
	case consts.RoleWp:
		nicid = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		nicid = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		nicid = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId
	case consts.RoleDs:
		nicid = mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index]
	}

	if len(nicid) != 0 {
		log.Print(awsCtx, "skipped already created the network interface", "id", nicid)
		return nicid, nil
	}

	interfaceparameter := &ec2.CreateNetworkInterfaceInput{
		Description: aws.String("network interface"),
		SubnetId:    aws.String(mainStateDocument.CloudInfra.Aws.SubnetIDs[0]),
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

	nicresponse, err := obj.client.BeginCreateNIC(ctx, interfaceparameter)
	if err != nil {
		return "", err
	}

	var errCreate error
	done := make(chan struct{})
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreate = err
		}
	}()
	<-done
	if errCreate != nil {
		return "", errCreate
	}
	log.Success(awsCtx, "Created network interface", "id", *nicresponse.NetworkInterface.NetworkInterfaceId)
	return *nicresponse.NetworkInterface.NetworkInterfaceId, nil
}

func (obj *AwsProvider) NewVM(storage ksctlTypes.StorageFactory, index int) error {
	name := <-obj.chResName
	indexNo := index
	role := <-obj.chRole
	vmtype := <-obj.chVMType

	log.Debug(awsCtx, "Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	instanceId := ""
	instanceIp := ""
	switch role {
	case consts.RoleCp:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo]
		instanceIp = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo]
	case consts.RoleDs:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo]
		instanceIp = mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo]
	case consts.RoleLb:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID
		instanceIp = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP
	case consts.RoleWp:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo]
		instanceIp = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo]
	}
	if len(instanceId) != 0 {
		if len(instanceIp) != 0 {
			log.Print(awsCtx, "skipped vm already created", "name", instanceId)
			return nil
		} else {
			instance_ip, err := obj.client.DescribeInstanceState(awsCtx, instanceId)
			if err != nil {
				return err
			}

			publicip := instance_ip.Reservations[0].Instances[0].PublicIpAddress
			privateip := instance_ip.Reservations[0].Instances[0].PrivateIpAddress

			obj.mu.Lock()
			defer obj.mu.Unlock()

			switch role {
			case consts.RoleWp:
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo] = *publicip
				mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP = *publicip
				mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PrivateIP = *privateip
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo] = *publicip
				mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs[indexNo] = *privateip
			}
		}
	}

	nicid, err := obj.CreateNetworkInterface(awsCtx, storage, name, indexNo, role)
	if err != nil {
		return err
	}

	ami, err := obj.getLatestUbuntuAMI()
	if err != nil {
		return err
	}
	initScript, err := helpers.GenerateInitScriptForVM(name)
	if err != nil {
		return err
	}
	initScriptBase64 := base64.StdEncoding.EncodeToString([]byte(initScript))

	parameter := &ec2.RunInstancesInput{
		ImageId:      aws.String(ami),
		InstanceType: types.InstanceType(vmtype),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String(mainStateDocument.CloudInfra.Aws.B.SSHKeyName),

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

	instanceop, err := obj.client.BeginCreateVM(awsCtx, parameter)
	if err != nil {
		return err
	}

	instanceId = *instanceop.Instances[0].InstanceId

	var errCreateVM error

	done1 := make(chan struct{})
	go func() {
		defer close(done1)
		obj.mu.Lock()
		defer obj.mu.Unlock()
		switch role {
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = instanceId
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.HostNames[indexNo] = name
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.VMSizes[indexNo] = vmtype
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = instanceId
			mainStateDocument.CloudInfra.Aws.InfoDatabase.HostNames[indexNo] = name
			mainStateDocument.CloudInfra.Aws.InfoDatabase.VMSizes[indexNo] = vmtype
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID = instanceId
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.HostName = name
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.VMSize = vmtype
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = instanceId
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.HostNames[indexNo] = name
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.VMSizes[indexNo] = vmtype
		}

		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
		}
	}()
	<-done1
	if errCreateVM != nil {
		return errCreateVM
	}

	log.Print(awsCtx, "creating vm", "vmName", name)

	done := make(chan struct{})

	go func() {
		defer close(done)

		err = obj.client.InstanceInitialWaiter(awsCtx, instanceId)
		if err != nil {
			errCreateVM = err
			return
		}

		instance_ip, err := obj.client.DescribeInstanceState(awsCtx, instanceId)
		if err != nil {
			errCreateVM = err
			return
		}

		publicip := instance_ip.Reservations[0].Instances[0].PublicIpAddress
		privateip := instance_ip.Reservations[0].Instances[0].PrivateIpAddress

		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo] = *publicip
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP = *publicip
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PrivateIP = *privateip
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PublicIPs[indexNo] = *publicip
			mainStateDocument.CloudInfra.Aws.InfoDatabase.PrivateIPs[indexNo] = *privateip
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreateVM = err
			return
		}
	}()
	<-done
	if errCreateVM != nil {
		return errCreateVM
	}

	log.Success(awsCtx, "Created the vm", "name", name)
	return nil
}

func (obj *AwsProvider) getLatestUbuntuAMI() (string, error) {
	imageFilter := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
			{
				Name:   aws.String("owner-alias"),
				Values: []string{"amazon"},
			},
		},
	}

	resp, err := obj.client.FetchLatestAMIWithFilter(imageFilter)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (obj *AwsProvider) DeleteNetworkInterface(ctx context.Context, storage ksctlTypes.StorageFactory, index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		interfaceName = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		interfaceName = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId
	case consts.RoleDs:
		interfaceName = mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index]
	}
	if len(interfaceName) == 0 {
		log.Print(awsCtx, "skipped already deleted the network interface")
	} else {
		err := obj.client.BeginDeleteNIC(interfaceName)
		if err != nil {
			return err
		}
		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkInterfaceId = ""
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkInterfaceIDs[index] = ""
		}
		err = storage.Write(mainStateDocument)
		if err != nil {
			return err
		}
		log.Success(awsCtx, "deleted the network interface", "id", interfaceName)
	}

	return nil
}

func fetchgroupid(role consts.KsctlRole) (string, error) {
	switch role {
	case consts.RoleCp:
		return mainStateDocument.CloudInfra.Aws.InfoControlPlanes.NetworkSecurityGroupIDs, nil
	case consts.RoleWp:
		return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroupIDs, nil
	case consts.RoleLb:
		return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroupID, nil
	case consts.RoleDs:
		return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroupIDs, nil

	}

	return "", ksctlErrors.ErrInvalidKsctlRole.Wrap(
		log.NewError(awsCtx, "invalid role", "role", role),
	)
}
