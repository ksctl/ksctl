package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"strconv"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

func (obj *AwsProvider) ec2Client() *ec2.Client {
	ec2client := ec2.NewFromConfig(obj.session)
	//TODO ADD ERROR HANDLING
	return ec2client
}

func (obj *AwsProvider) vpcClienet() ec2.CreateVpcInput {

	vpcClient := ec2.CreateVpcInput{
		CidrBlock: aws.String("172.31.0.0/16"),
	}
	fmt.Println("VPC Client Created Successfully")
	return vpcClient
}

func (obj *AwsProvider) CreateVPC() {

	vpcClient := obj.vpcClienet()
	ec2Client := obj.ec2Client()

	vpc, err := ec2Client.CreateVpc(context.TODO(), &vpcClient)
	if err != nil {
		log.Error("Error Creating VPC", err)
	}
	awsCloudState.VPCID = *vpc.Vpc.VpcId
	fmt.Print("VPC Created Successfully: ")
	fmt.Println(*vpc.Vpc.VpcId)

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
		VpcId:             aws.String(VPCID),
	})
	if err != nil {
		log.Error("Error Attaching Internet Gateway", err)
	}
	GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	fmt.Println(*createInternetGateway.InternetGateway.InternetGatewayId)
	awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
	fmt.Print("Internet Gateway Created Successfully: ")

	return nil
}

func (obj *AwsProvider) CreateRouteTable() {

	ec2Client := obj.ec2Client()

	routeTableClient := ec2.CreateRouteTableInput{
		VpcId: aws.String(VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("route-table"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("DEMO" + "-rt"),
					},
				},
			},
		},
	}

	routeTable, err := ec2Client.CreateRouteTable(context.TODO(), &routeTableClient)
	if err != nil {
		log.Error("Error Creating Route Table", err)
	}

	log.Success("Route Table Created Successfully ", "id: ", *routeTable.RouteTable.RouteTableId)
	RouteTableID = *routeTable.RouteTable.RouteTableId

	for _, subnet := range SUBNETID {
		_, err = ec2Client.AssociateRouteTable(context.TODO(), &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
			SubnetId:     aws.String(subnet),
		})
		if err != nil {
			log.Error("Error Associating Route Table", err)
		}
	}

	fmt.Println("Route Table Associated Successfully....")
	/*        create route		*/
	_, err = ec2Client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(GatewayID),
		RouteTableId:         aws.String(*routeTable.RouteTable.RouteTableId),
	})
	if err != nil {
		log.Error("Error Creating Route", err)
	}

	log.Success("Route Table Created Successfully: ", "id: ", *routeTable.RouteTable.RouteTableId)

}

func (obj *AwsProvider) ElbClient() *elasticloadbalancingv2.Client {
	elbv2Client := elasticloadbalancingv2.NewFromConfig(obj.session)

	return elbv2Client
}

var (
	GLBARN *elasticloadbalancingv2.CreateLoadBalancerOutput
	GARN   *elasticloadbalancingv2.CreateTargetGroupOutput
)

//// TODO : NOT CONFIRMED
//func (obj *AwsProvider) CreateLB() (*elasticloadbalancingv2.CreateLoadBalancerOutput, error) {
//
//	LBCLIENT := obj.ElbClient()
//	LB_ARN, err := LBCLIENT.CreateLoadBalancer(context.TODO(), &elasticloadbalancingv2.CreateLoadBalancerInput{
//		Name:           aws.String("new" + "-lb"),
//		Scheme:         elb_types.LoadBalancerSchemeEnumInternetFacing,
//		IpAddressType:  elb_types.IpAddressType("ipv4"),
//		SecurityGroups: []string{awsCloudState.SecurityGroupID},
//		Subnets:        []string{awsCloudState.SubnetID},
//		Type:           elb_types.LoadBalancerTypeEnumApplication,
//	})
//	if err != nil {
//		log.Println(err)
//	}
//	GLBARN = LB_ARN
//	return LB_ARN, nil
//
//}

func (obj *AwsProvider) CreateTargetGroup() (*elasticloadbalancingv2.CreateTargetGroupOutput, error) {

	client := obj.ElbClient()

	ARN, err := client.CreateTargetGroup(context.TODO(), &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String("new" + "-tg"),
		Protocol:   elb_types.ProtocolEnumTcp,
		Port:       aws.Int32(6443),
		VpcId:      aws.String(VPCID),
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
				// TODO: Add the more parameters
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
							// Weight:         aws.Int32(1),
						},
					},
				},
			},
		},
		LoadBalancerArn: aws.String(*GLBARN.LoadBalancers[0].LoadBalancerArn),
		Port:            aws.Int32(6443), // port on which the load balancer listens
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

// TODO add EBS volume to the VM and attach it to the instance

func (obj *AwsProvider) DelVM(storage resources.StorageFactory, index int) error {

	role := obj.metadata.role
	indexNo := index
	obj.mxRole.Unlock()

	vmName := ""

	switch role {
	case consts.RoleCp:
		vmName = awsCloudState.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = awsCloudState.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = awsCloudState.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = awsCloudState.InfoWorkerPlanes.Names[indexNo]
	}

	if len(vmName) == 0 {
		log.Success("[skip] already deleted the vm", "vmname", vmName)
	} else {

		var errDel error
		donePoll := make(chan struct{})
		go func() {
			defer close(donePoll)

			log.Print("deleting vm...", "name", vmName)

			err := obj.client.BeginDeleteVM(vmName, obj.ec2Client())
			if err != nil {
				errDel = err
				return
			}
			switch role {
			case consts.RoleWp:
				awsCloudState.InfoWorkerPlanes.Names[indexNo] = ""
			case consts.RoleCp:
				awsCloudState.InfoControlPlanes.Names[indexNo] = ""
			case consts.RoleLb:
				awsCloudState.InfoLoadBalancer.Name = ""
			case consts.RoleDs:
				awsCloudState.InfoDatabase.Names[indexNo] = ""
			}

			err = obj.DeleteNetworkInterface(context.Background(), storage, indexNo, role)
			if err != nil {
				errDel = err
				return
			}
			obj.mxState.Lock()
			defer obj.mxState.Unlock()

			if err := saveStateHelper(storage); err != nil {
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

	securitygroup, err := fetchgroupid(role)

	nicid := ""
	switch role {
	case consts.RoleWp:
		nicid = awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		nicid = awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		nicid = awsCloudState.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		nicid = awsCloudState.InfoDatabase.NetworkInterfaceIDs[index]
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
		SubnetId:    aws.String(awsCloudState.SubnetID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("network-interface"),
				Tags: []types.Tag{
					{
						Key:   aws.String(obj.metadata.resName),
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
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
			//awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames = append(awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames, *nicresponse.NetworkInterface.NetworkInterfaceId)
		case consts.RoleCp:
			awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
			//awsCloudState.InfoControlPlanes.NetworkInterfaceNames = append(awsCloudState.InfoControlPlanes.NetworkInterfaceNames, *nicresponse.NetworkInterface.NetworkInterfaceId)
		case consts.RoleLb:
			awsCloudState.InfoLoadBalancer.NetworkInterfaceName = *nicresponse.NetworkInterface.NetworkInterfaceId
		case consts.RoleDs:
			awsCloudState.InfoDatabase.NetworkInterfaceIDs[index] = *nicresponse.NetworkInterface.NetworkInterfaceId
			//awsCloudState.InfoDatabase.NetworkInterfaceNames = append(awsCloudState.InfoDatabase.NetworkInterfaceNames, *nicresponse.NetworkInterface.NetworkInterfaceId)

		default:
			errCreate = fmt.Errorf("invalid role %s", role)
		}
	}()

	if errCreate != nil {
		fmt.Println(errCreate)
		return "", errCreate
	}
	log.Success("[aws] created the network interface ", "id: ", *nicresponse.NetworkInterface.NetworkInterfaceId)

	return *nicresponse.NetworkInterface.NetworkInterfaceId, nil
}

func (obj *AwsProvider) NewVM(storage resources.StorageFactory, indexNo int) error {
	role := obj.metadata.role
	vmtype := obj.metadata.vmType

	obj.mxRole.Unlock()
	obj.mxName.Unlock()
	obj.mxVMType.Unlock()

	if obj.metadata.role == consts.RoleDs && indexNo > 0 {
		log.Note("[skip] currently multiple datastore not supported")
		return nil
	}

	vmName := ""
	switch role {
	case consts.RoleCp:
		vmName = awsCloudState.InfoControlPlanes.Names[indexNo]
	case consts.RoleDs:
		vmName = awsCloudState.InfoDatabase.Names[indexNo]
	case consts.RoleLb:
		vmName = awsCloudState.InfoLoadBalancer.Name
	case consts.RoleWp:
		vmName = awsCloudState.InfoWorkerPlanes.Names[indexNo]
	}
	if len(vmName) != 0 {
		log.Print("skipped vm already created", "name", vmName)
		return nil
	}

	stringindexNo := fmt.Sprintf("%d", indexNo)
	ec2Client := obj.ec2Client()

	nicid, err := obj.CreateNetworkInterface(context.TODO(), storage, obj.metadata.resName, indexNo, obj.metadata.role)
	if err != nil {
		panic("Error creating network interface: " + err.Error())
	}

	parameter := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0287a05f0ef0e9d9a"),
		InstanceType: types.InstanceType(vmtype),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String("test-e2e-ha-aws-ssh"),
		Monitoring: &types.RunInstancesMonitoringEnabled{
			Enabled: aws.Bool(true),
		},

		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Arn: aws.String("arn:aws:iam::708808958753:instance-profile/kssctl-arn"),
		},

		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("instance"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(string(role) + stringindexNo),
					},
					{
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},

		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:        aws.Int32(0),
				NetworkInterfaceId: aws.String(nicid),
				// AssociatePublicIpAddress: aws.Bool(true),
			},
		},
	}

	instanceop, err := obj.client.BeginCreateVM(context.Background(), ec2Client, parameter)
	if err != nil {
		log.Error("Error creating vm", "error", err)
	}

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

	var errCreate error

	done := make(chan struct{})
	go func() {
		defer close(done)
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case consts.RoleWp:
			awsCloudState.InfoWorkerPlanes.Names[indexNo] = *instanceop.Instances[0].InstanceId
			awsCloudState.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
			awsCloudState.InfoWorkerPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleCp:
			awsCloudState.InfoControlPlanes.Names[indexNo] = *instanceop.Instances[0].InstanceId
			awsCloudState.InfoControlPlanes.PublicIPs[indexNo] = *publicip
			awsCloudState.InfoControlPlanes.PrivateIPs[indexNo] = *privateip
		case consts.RoleLb:
			awsCloudState.InfoLoadBalancer.Name = *instanceop.Instances[0].InstanceId
			awsCloudState.InfoLoadBalancer.PublicIP = *publicip
			awsCloudState.InfoLoadBalancer.PrivateIP = *privateip
		case consts.RoleDs:
			awsCloudState.InfoDatabase.Names[indexNo] = *instanceop.Instances[0].InstanceId
			awsCloudState.InfoDatabase.PublicIPs[indexNo] = *publicip
			awsCloudState.InfoDatabase.PrivateIPs[indexNo] = *privateip
		}
		if err := saveStateHelper(storage); err != nil {
			errCreate = err
			fmt.Println(err)
			return
		}
	}()
	<-done
	if errCreate != nil {
		fmt.Println(errCreate)
		return errCreate
	}

	log.Success("Created virtual machine", "id: ", *instanceop.Instances[0].InstanceId)
	return nil
}

func (obj *AwsProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role consts.KsctlRole) error {

	interfaceName := ""
	switch role {
	case consts.RoleWp:
		interfaceName = awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index]
	case consts.RoleCp:
		interfaceName = awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[index]
	case consts.RoleLb:
		interfaceName = awsCloudState.InfoLoadBalancer.NetworkInterfaceName
	case consts.RoleDs:
		interfaceName = awsCloudState.InfoDatabase.NetworkInterfaceIDs[index]
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
			awsCloudState.InfoWorkerPlanes.NetworkInterfaceIDs[index] = ""
			//awsCloudState.InfoWorkerPlanes.NetworkInterfaceNames[index] = ""
		case consts.RoleCp:
			awsCloudState.InfoControlPlanes.NetworkInterfaceIDs[index] = ""
			//awsCloudState.InfoControlPlanes.NetworkInterfaceNames[index] = ""
		case consts.RoleLb:
			awsCloudState.InfoLoadBalancer.NetworkInterfaceName = ""
		case consts.RoleDs:
			awsCloudState.InfoDatabase.NetworkInterfaceIDs[index] = ""
			//awsCloudState.InfoDatabase.NetworkInterfaceNames[index] = ""
		default:
			return fmt.Errorf("invalid role %s", role)
		}
		err = saveStateHelper(storage)
		if err != nil {
			log.Error("Error saving state", "error", err)
		}
		log.Success("[aws] deleted the network interface ", "id: ", interfaceName)
	}

	return nil
}

func (obj *AwsProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role string) error {
	return nil
}

func (obj *AwsProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role string, instancid string) (*ec2.AllocateAddressOutput, error) {

	ec2Client := obj.ec2Client()
	// now we will assign the public ip to the instance
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
		return awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup, nil
	case consts.RoleWp:
		return awsCloudState.InfoWorkerPlanes.NetworkSecurityGroup, nil
	case consts.RoleLb:
		return awsCloudState.InfoLoadBalancer.NetworkSecurityGroup, nil
	case consts.RoleDs:
		return awsCloudState.InfoDatabase.NetworkSecurityGroup, nil

	}

	return "", fmt.Errorf("invalid role %s", role)
}
