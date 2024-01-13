package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func (obj *AwsProvider) ec2Client() *ec2.Client {
	ec2client := ec2.NewFromConfig(obj.session)
	return ec2client
}

func (obj *AwsProvider) CreateInternetGateway() error {

	ec2Client := obj.ec2Client()

	internetGateway := ec2.CreateInternetGatewayInput{
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("internet-gateway"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("DEMO" + "-ig"),
					},
				},
			},
		},
	}

	createInternetGateway, err := ec2Client.CreateInternetGateway(context.TODO(), &internetGateway)
	if err != nil {
		log.Error("Error Creating Internet Gateway", err)
	}

	_, err = ec2Client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(mainStateDocument.CloudInfra.Aws.VPCID),
	})
	if err != nil {
		log.Error("Error Attaching Internet Gateway", err)
	}

	fmt.Println(*createInternetGateway.InternetGateway.InternetGatewayId)
	mainStateDocument.CloudInfra.Aws.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	fmt.Print("Internet Gateway Created Successfully: ")

	return nil
}

func (obj *AwsProvider) ElbClient() *elasticloadbalancingv2.Client {
	elbv2Client := elasticloadbalancingv2.NewFromConfig(obj.session)

	return elbv2Client
}

var (
	GLBARN *elasticloadbalancingv2.CreateLoadBalancerOutput
	GARN   *elasticloadbalancingv2.CreateTargetGroupOutput
)

func (obj *AwsProvider) CreateTargetGroup() (*elasticloadbalancingv2.CreateTargetGroupOutput, error) {

	client := obj.ElbClient()

	ARN, err := client.CreateTargetGroup(context.TODO(), &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String("new" + "-tg"),
		Protocol:   elb_types.ProtocolEnumTcp,
		Port:       aws.Int32(6443),
		VpcId:      aws.String(mainStateDocument.CloudInfra.Aws.VPCID),
		TargetType: elb_types.TargetTypeEnumIp,
	})

	if err != nil {
		log.Error("Error Creating Target Group", err)
	}
	GARN = ARN
	fmt.Println("Target Group Created Successfully: ", *ARN.TargetGroups[0].TargetGroupArn)

	return ARN, nil

}

func (obj *AwsProvider) RegisterTargetGroup() {
	client := obj.ElbClient()

	ARN := GARN
	if ARN == nil {
		log.Error("Target Group not created")
		return
	}

	ARNV := ARN.TargetGroups[0].TargetGroupArn

	_, err := client.RegisterTargets(context.TODO(), &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(*ARNV),
		Targets: []elb_types.TargetDescription{
			{
				Id: aws.String(""),
			},
		},
	})
	if err != nil {
		log.Error("Error Registering Target Group", err)
	}

	fmt.Println("Target Group Registered Successfully: ", *GLBARN.LoadBalancers[0].LoadBalancerArn)
}

func (obj *AwsProvider) CreateListener() {
	client := obj.ElbClient()

	ARN := GARN

	ARNV := ARN.TargetGroups[0].TargetGroupArn

	client.CreateListener(context.TODO(), &elasticloadbalancingv2.CreateListenerInput{
		DefaultActions: []elb_types.Action{
			{
				Type: elb_types.ActionTypeEnumForward,
				ForwardConfig: &elb_types.ForwardActionConfig{
					TargetGroups: []elb_types.TargetGroupTuple{
						{
							TargetGroupArn: aws.String(*ARNV),
						},
					},
				},
			},
		},
		LoadBalancerArn: aws.String(*GLBARN.LoadBalancers[0].LoadBalancerArn),
		Port:            aws.Int32(6443),
		Protocol:        elb_types.ProtocolEnumTcp,
		Tags: []elb_types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("new" + "-listener"),
			},
		},
	})

	log.Success("Listener Created Successfully: ", "id", *GLBARN.LoadBalancers[0].LoadBalancerArn)

}

func (obj *AwsProvider) DelVM(storage resources.StorageFactory, index int) error {

	role := <-obj.chRole
	indexNo := index

	log.Debug("Printing", "role", role, "indexNo", indexNo)

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
		log.Success("[skip] already deleted the vm", "vmname", vmName)
	} else {

		var errDel error
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)

			err := obj.client.BeginDeleteVM(vmName, obj.ec2Client())
			if err != nil {
				errDel = err
				return
			}
			obj.mu.Lock()
			defer obj.mu.Unlock()

			switch role {
			case consts.RoleWp:
				mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = ""
			case consts.RoleCp:
				mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = ""
			case consts.RoleLb:
				mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID = ""
			case consts.RoleDs:
				mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = ""
			}

			err = obj.DeleteNetworkInterface(context.Background(), storage, indexNo, role)
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
			return log.NewError(errDel.Error())
		}
		log.Success("Deleted the vm", "id: ", vmName)

	}
	return nil
}

func (obj *AwsProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory, resName string, index int, role consts.KsctlRole) (string, error) {
	obj.nicprint.Do(func() {
		log.Print("Creating the network interface...")
	})

	securitygroup, err := fetchgroupid(role)
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
		log.Print("[skip] already created the network interface", "id", nicid)
		return nicid, nil
	}

	if err != nil {
		log.Error("Error fetching security group id", "error", err)
	}

	interfaceparameter := &ec2.CreateNetworkInterfaceInput{
		Description: aws.String("network interface"),
		SubnetId:    aws.String(mainStateDocument.CloudInfra.Aws.SubnetID),
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

	nicresponse, err := obj.client.BeginCreateNIC(ctx, obj.ec2Client(), interfaceparameter)
	if err != nil {
		log.Error("Error creating network interface", "error", err)
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

		default:
			errCreate = fmt.Errorf("invalid role %s", role)
		}
		if err := storage.Write(mainStateDocument); err != nil {
			errCreate = err
		}
	}()
	<-done
	if errCreate != nil {
		return "", errCreate
	}
	log.Print("Created network interface", "id", *nicresponse.NetworkInterface.NetworkInterfaceId)
	return *nicresponse.NetworkInterface.NetworkInterfaceId, nil
}

func (obj *AwsProvider) NewVM(storage resources.StorageFactory, index int) error {
	name := <-obj.chResName
	indexNo := index
	role := <-obj.chRole
	vmtype := <-obj.chVMType

	log.Debug("Printing", "name", name, "indexNo", indexNo, "role", role, "vmType", vmtype)

	if role == consts.RoleDs && indexNo > 0 {
		log.Print("skipped currently multiple datastore not supported")
		return nil
	}

	instanceId := ""
	switch role {
	case consts.RoleCp:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo]
	case consts.RoleDs:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo]
	case consts.RoleLb:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID
	case consts.RoleWp:
		instanceId = mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo]
	}
	if len(instanceId) != 0 {
		log.Print("skipped vm already created", "name", instanceId)
		return nil
	}

	stringindexNo := fmt.Sprintf("%d", indexNo)
	ec2Client := obj.ec2Client()

	nicid, err := obj.CreateNetworkInterface(context.TODO(), storage, name, indexNo, role)
	if err != nil {
		log.NewError("Error creating network interface", "error", err)
	}

	ami, err := getLatestUbuntuAMI(ec2Client)
	if err != nil {
		log.Error("Error getting latest ubuntu ami", "error", err)
	}
	initScript, err := helpers.GenerateInitScriptForVM(name)
	if err != nil {
		log.Error("Error generating init script", "error", err)
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
						Value: aws.String(string(role) + stringindexNo),
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

	instanceop, err := obj.client.BeginCreateVM(context.Background(), ec2Client, parameter)
	if err != nil {
		log.Error("Error creating vm", "error", err)
	}

	instanceipinput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceop.Instances[0].InstanceId},
	}

	instance_ip, err := obj.client.DescribeInstanceState(context.Background(), ec2Client, instanceipinput)
	if err != nil {
		log.Error("Error getting instance state", "error", err)
	}

	publicip := instance_ip.Reservations[0].Instances[0].PublicIpAddress
	privateip := instance_ip.Reservations[0].Instances[0].PrivateIpAddress

	var errCreateVM error

	done := make(chan struct{})
	go func() {
		defer close(done)
		obj.mu.Lock()
		defer obj.mu.Unlock()

		switch role {
		case consts.RoleWp:
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.InstanceIds[indexNo] = *instanceop.Instances[0].InstanceId
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.Names[indexNo] = name
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
			mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleCp:
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.InstanceIds[indexNo] = *instanceop.Instances[0].InstanceId
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.Names[indexNo] = name
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PublicIPs[indexNo] = *publicip
			mainStateDocument.CloudInfra.Aws.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleLb:
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.InstanceID = *instanceop.Instances[0].InstanceId
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.Name = name
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PublicIP = *publicip
			mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.PrivateIP = *privateip
		case consts.RoleDs:
			mainStateDocument.CloudInfra.Aws.InfoDatabase.InstanceIds[indexNo] = *instanceop.Instances[0].InstanceId
			mainStateDocument.CloudInfra.Aws.InfoDatabase.Names[indexNo] = name
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
		return log.NewError(errCreateVM.Error())
	}

	log.Debug("Printing", "mainStateDocument", mainStateDocument)

	log.Success("Created the vm", "id: ", *instanceop.Instances[0].InstanceId)
	return nil
}

func getLatestUbuntuAMI(client *ec2.Client) (string, error) {
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

	resp, err := client.DescribeImages(context.TODO(), imageFilter)
	if err != nil {
		return "", fmt.Errorf("failed to describe images: %w", err)
	}
	if len(resp.Images) == 0 {
		return "", fmt.Errorf("no images found")
	}

	var savedImages []types.Image

	for _, i := range resp.Images {
		if trustedSource(*i.OwnerId) && *i.Public {
			savedImages = append(savedImages, i)
		}
	}

	sort.Slice(savedImages, func(i, j int) bool {
		return *savedImages[i].CreationDate > *savedImages[j].CreationDate
	})

	for x := 0; x < 2; x++ {
		i := savedImages[x]
		if i.ImageOwnerAlias != nil {
			log.Debug("ownerAlias", *i.ImageOwnerAlias)
		}
		log.Debug("Printing amis", "creationdate", *i.CreationDate, "public", *i.Public, "ownerid", *i.OwnerId, "architecture", i.Architecture.Values(), "name", *i.Name, "imageid", *i.ImageId)
	}

	selectedAMI := *savedImages[0].ImageId

	log.Print("Selected ami image", "imageId", selectedAMI)

	return selectedAMI, nil
}

// trustedSource: helper recieved from https://ubuntu.com/tutorials/search-and-launch-ubuntu-22-04-in-aws-using-cli#2-search-for-the-right-ami
func trustedSource(id string) bool {
	// 679593333241
	// 099720109477
	if strings.Compare(id, "679593333241") != 0 && strings.Compare(id, "099720109477") != 0 {
		return false
	}
	return true
}

func (obj *AwsProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {

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
		log.Print("skipped network interface already deleted")
	} else {
		err := obj.client.BeginDeleteNIC(interfaceName, obj.ec2Client())
		if err != nil {
			log.Error("Error deleting network interface", "error", err)
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
		default:
			return fmt.Errorf("invalid role %s", role)
		}
		err = storage.Write(mainStateDocument)
		if err != nil {
			log.Error("Error saving state", "error", err)
		}
		log.Success("[aws] deleted the network interface ", "id: ", interfaceName)
	}

	return nil
}

// TODO : NOT CONFIRMED
func (obj *AwsProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role string) error {
	return nil
}

// TODO : NOT CONFIRMED
func (obj *AwsProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role string, instancid string) (*ec2.AllocateAddressOutput, error) {

	ec2Client := obj.ec2Client()
	allocRes, err := ec2Client.AllocateAddress(context.Background(), &ec2.AllocateAddressInput{
		Domain: types.DomainType("vpc"),
	})
	if err != nil {
		log.Error("Error Creating Public IP", err)
	}

	_, err = ec2Client.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		InstanceId:         aws.String(instancid),
		AllocationId:       allocRes.AllocationId,
		AllowReassociation: aws.Bool(true),
	})

	log.Success("[aws] created the public IP ", ":", *allocRes.PublicIp)
	log.Success("[aws] attached the public IP %s to the instance %s", "id: ", *allocRes.PublicIp, instancid)
	return nil, nil
}

func (obj *AwsProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role string) error {

	return nil
}

func fetchgroupid(role consts.KsctlRole) (string, error) {
	switch role {
	case consts.RoleCp:
		return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup, nil
	case consts.RoleWp:
		return mainStateDocument.CloudInfra.Aws.InfoWorkerPlanes.NetworkSecurityGroup, nil
	case consts.RoleLb:
		return mainStateDocument.CloudInfra.Aws.InfoLoadBalancer.NetworkSecurityGroup, nil
	case consts.RoleDs:
		return mainStateDocument.CloudInfra.Aws.InfoDatabase.NetworkSecurityGroup, nil

	}

	return "", fmt.Errorf("invalid role %s", role)
}
