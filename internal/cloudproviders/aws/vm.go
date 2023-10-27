package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/kubesimplify/ksctl/pkg/resources"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
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
		fmt.Println("Error Creating VPC")
		log.Println(err)
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
		log.Println(err)
	}

	_, err = ec2Client.AttachInternetGateway(context.TODO(), &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*createInternetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(VPCID),
	})
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}

	fmt.Print("Route Table Created Successfully: ")
	fmt.Println(*routeTable.RouteTable.RouteTableId)
	RouteTableID = *routeTable.RouteTable.RouteTableId

	for _, subnet := range SUBNETID {
		_, err = ec2Client.AssociateRouteTable(context.TODO(), &ec2.AssociateRouteTableInput{
			RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
			SubnetId:     aws.String(subnet),
		})
		if err != nil {
			log.Println(err)
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
		log.Println(err)
	}

	fmt.Println("Route Created Successfully....")

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
		log.Println(err)
	}
	GARN = ARN
	fmt.Println("Target Group Created Successfully: ", *ARN.TargetGroups[0].TargetGroupArn)

	return ARN, nil

}

func (obj *AwsProvider) RegisterTargetGroup() {
	client := obj.ElbClient()

	ARN := GARN
	if ARN == nil {
		log.Println("Target Group not created")
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
		fmt.Println("Could not register the target group")
		log.Fatal(err)
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

	fmt.Println("Listener Created Successfully: ", *GLBARN.LoadBalancers[0].LoadBalancerArn)

}

// TODO add EBS volume to the VM and attach it to the instance

// TODO ADD A GLOBAL FUNTION THAT WILL HAVE THE ALL OUTPUTS

// Sequence of steps to create a VM
// 1. Create  VPC										DONE
// 2. Create  Subnet									DONE
// 3. Create  Internet Gateway							DONE
// 4. Create  Route Table								DONE
// 5. Create  Firewall aka Security Group in AWS		DONE
// 6. Create Load Balancer								DONE
// 7. Create Public IP									DONE
// 8. OS IAMGE											DONE
// 9. Generate SSH Key
// 10. Create VM

// TODO Refactor all the code same as various providor

func (obj *AwsProvider) DelVM(factory resources.StorageFactory, index int) error {

	//role := obj.metadata.role
	//indexNo := index
	//obj.mxRole.Unlock()
	//
	////ec2Client := obj.ec2Client()
	//
	////vmName := ""
	//switch role {
	//case ROLE_CP:
	//	vmName = awsCloudState.InfoControlPlanes.Names[indexNo]
	//case ROLE_DS:
	//	vmName = awsCloudState.InfoDatabase.Names[indexNo]
	//case ROLE_LB:
	//	vmName = awsCloudState.InfoLoadBalancer.Name
	//case ROLE_WP:
	//	vmName = awsCloudState.InfoWorkerPlanes.Names[indexNo]
	//}

	return nil
}

var NICID string

func (obj *AwsProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory, resName string, index int, role KsctlRole) (*ec2.CreateNetworkInterfaceOutput, error) {

	groupid_role, err := fetchgroupid(role)

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
			groupid_role,
		},
	}

	vniclient := obj.ec2Client()
	nicresponse, err := vniclient.CreateNetworkInterface(ctx, interfaceparameter)
	if err != nil {
		log.Println(err)
	}

	NICID = *nicresponse.NetworkInterface.NetworkInterfaceId
	// wait until the instance state comes from pending to running use
	// time.Sleep(10 * time.Second)

	storage.Logger().Success("[aws] created the network interface ", *nicresponse.NetworkInterface.NetworkInterfaceId)

	return nicresponse, nil

}

func (obj *AwsProvider) NewVM(storage resources.StorageFactory, indexNo int) error {

	//name := obj.metadata.resName
	role := obj.metadata.role
	//vmtype := obj.metadata.vmType

	obj.mxRole.Unlock()
	obj.mxName.Unlock()

	if obj.metadata.role == RoleDs && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}

	stringindexNo := fmt.Sprintf("%d", indexNo)
	ec2Client := obj.ec2Client()

	_, err := obj.CreateNetworkInterface(context.TODO(), storage, obj.metadata.resName, indexNo, obj.metadata.role)
	if err != nil {
		panic("Error creating network interface: " + err.Error())
	}

	parameter := &ec2.RunInstancesInput{
		// use awslinux image id ---->   ami-0e306788ff2473ccb
		ImageId:      aws.String("ami-067c21fb1979f0b27"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String("testkeypair"),
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
				NetworkInterfaceId: aws.String(NICID),
				// AssociatePublicIpAddress: aws.Bool(true),
			},
		},
	}

	instanceop, err := ec2Client.RunInstances(context.Background(), parameter)
	if err != nil {
		fmt.Println(err)
		panic("Error creating EC2 instance: " + err.Error())
	}

	var state int32 = 0
	for {
		vmstate, err := ec2Client.DescribeInstanceStatus(context.Background(), &ec2.DescribeInstanceStatusInput{
			InstanceIds: []string{*instanceop.Instances[0].InstanceId},
		})
		if err != nil {
			log.Println(err)
		}
		state = *vmstate.InstanceStatuses[0].InstanceState.Code
		if state == 16 {
			storage.Logger().Success("[aws] instance running ")
			break
		}
	}
	if err != nil {
		return err
	}

	// get the instance public ip
	instanceip := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceop.Instances[0].InstanceId},
	}

	instance_ip, err := ec2Client.DescribeInstances(context.Background(), instanceip)
	if err != nil {
		return err
	}

	publicip := instance_ip.Reservations[0].Instances[0].PublicIpAddress
	fmt.Println(*publicip)

	// TODO : make sure err not fue to vm type mutex

	done := make(chan struct{})
	var errCreate error
	go func() {
		defer close(done)
		obj.mxState.Lock()
		defer obj.mxState.Unlock()

		switch role {
		case RoleWp:

			awsCloudState.InfoWorkerPlanes.PublicIPs[indexNo] = *publicip
		case RoleCp:
			awsCloudState.InfoControlPlanes.PublicIPs[indexNo] = *publicip
		case RoleLb:
			awsCloudState.InfoLoadBalancer.PublicIP = *publicip
		case RoleDs:
			awsCloudState.InfoDatabase.PublicIPs[indexNo] = *publicip
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

	storage.Logger().Success("[aws] created the instance ", *instanceop.Instances[0].InstanceId)
	return nil
}

func (obj *AwsProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role string) error {
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
		log.Println(err)
	}

	_, err = ec2Client.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		InstanceId:         aws.String(instancid),
		AllocationId:       allocRes.AllocationId,
		AllowReassociation: aws.Bool(true),
	})

	storage.Logger().Success("[aws] created the public IP ", *allocRes.PublicIp)
	storage.Logger().Success("[aws] attached the public IP %s to the instance %s", *allocRes.PublicIp, instancid)
	return nil, nil
}

func (obj *AwsProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role string) error {

	return nil
}

func fetchgroupid(role KsctlRole) (string, error) {
	switch role {
	case RoleCp:
		return awsCloudState.SecurityGroupID[0], nil
	case RoleWp:
		return awsCloudState.SecurityGroupID[1], nil
	case RoleLb:
		return awsCloudState.SecurityGroupID[2], nil
	case RoleDs:
		return awsCloudState.SecurityGroupID[3], nil

	}
	return "", fmt.Errorf("No security group found for role %s", role)
}
