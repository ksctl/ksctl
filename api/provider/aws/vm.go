package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

var (
	awsCloudState *StateConfiguration
	GatewayID     string
	RouteTableID  string
	VPCID         string
	SUBNETID      []string
)

func (obj *AwsProvider) ec2Client() *ec2.Client {
	ec2client := ec2.NewFromConfig(obj.session)
	//TODO ADD ERROR HANDLING
	return ec2client
}

func (obj *AwsProvider) vpcClienet() ec2.CreateVpcInput {

	vpcClient := ec2.CreateVpcInput{
		CidrBlock: aws.String("172.31.0.0/16"),
		// Dry run is used to check if the request is valid
		// without actually creating the VPC.
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
	VPCID = *vpc.Vpc.VpcId
	// awsCloudState.VPC = *vpc.Vpc.VpcId
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
	// awsCloudState.GatewayID = *createInternetGateway.InternetGateway.InternetGatewayId
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

/*
	create lb   				DONE
	create target group			DONE
	register target group		DONE
	create listener				DONE
*/

// TODO: Use elb v2 client
func (obj *AwsProvider) ElbClient() *elasticloadbalancingv2.Client {
	elbv2Client := elasticloadbalancingv2.NewFromConfig(obj.session)

	return elbv2Client
}

var (
	GLBARN *elasticloadbalancingv2.CreateLoadBalancerOutput
	GARN   *elasticloadbalancingv2.CreateTargetGroupOutput
)

func (obj *AwsProvider) CreateLB() (*elasticloadbalancingv2.CreateLoadBalancerOutput, error) {

	LBCLIENT := obj.ElbClient()
	LB_ARN, err := LBCLIENT.CreateLoadBalancer(context.TODO(), &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:           aws.String("new" + "-lb"),
		Scheme:         elb_types.LoadBalancerSchemeEnumInternetFacing,
		IpAddressType:  elb_types.IpAddressType("ipv4"),
		SecurityGroups: []string{awsCloudState.SecurityGroupID},
		Subnets:        []string{awsCloudState.SubnetID},
		Type:           elb_types.LoadBalancerTypeEnumApplication,
	})
	if err != nil {
		log.Println(err)
	}
	GLBARN = LB_ARN
	return LB_ARN, nil

}

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

func (obj *AwsProvider) PublicIP(storage resources.StorageFactory, publicIPName string, index int) error {

	publicIP := ""
	switch obj.metadata.role {
	case utils.ROLE_WP:
		publicIP = awsCloudState.InfoWorkerPlanes.PublicIPNames[index]
	case utils.ROLE_CP:
		publicIP = awsCloudState.InfoControlPlanes.PublicIPNames[index]
	case utils.ROLE_LB:
		publicIP = awsCloudState.InfoLoadBalancer.PublicIP
	case utils.ROLE_DS:
		publicIP = awsCloudState.InfoDatabase.PublicIPNames[index]
	}
	if len(publicIP) != 0 {
		storage.Logger().Success("[skip] pub ip already created", publicIP)
		return nil
	}

	client := obj.ec2Client()
	Ipv4Pool, err := client.CreatePublicIpv4Pool(context.TODO(), &ec2.CreatePublicIpv4PoolInput{
		TagSpecifications: []types.TagSpecification{
			{
				Tags: []types.Tag{
					{
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
	}

	parameters := &ec2.DescribeAddressesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("domain"),
				Values: []string{"vpc"},
			},
		},
		AllocationIds: []string{
			*Ipv4Pool.PoolId,
		},
	}
	ipaddress, err := client.DescribeAddresses(context.Background(), parameters)
	if err != nil {
		log.Println(err)
	}
	if ipaddress.Addresses == nil {
		fmt.Printf("No elastic IPs for %s region\n", obj.region)
		return err
	}
	fmt.Println("Elastic IPs")
	for _, addr := range ipaddress.Addresses {
		fmt.Println("*", fmtAddress(&addr))
	}

	// TODO its just a pool create the public ip
	allocateIp, err := client.AllocateAddress(context.Background(), &ec2.AllocateAddressInput{
		Domain:                types.DomainType("vpc"),
		CustomerOwnedIpv4Pool: aws.String(*Ipv4Pool.PoolId),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("elastic-ip"),
				Tags: []types.Tag{
					{
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
	}
	if allocateIp.CustomerOwnedIpv4Pool != aws.String(*Ipv4Pool.PoolId) {
		fmt.Println("Elastic IP is not allocated from the pool")
		return nil
	}
	fmt.Println("Public IP Created Successfully: ", allocateIp.PublicIp)
	return nil
}

func (obj *AwsProvider) AssignPublicIP(instanceid string, publicip string) error {
	client := obj.ec2Client()
	_, err := client.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		InstanceId: aws.String(instanceid),
		PublicIp:   aws.String(publicip),
	})
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Public IP %s is assigned to %s\n", "", instanceid)

	return nil
}

func fmtAddress(addr *types.Address) string {
	out := fmt.Sprintf("IP: %s,  allocation id: %s",
		*aws.String(*addr.PublicIp), *aws.String(*addr.PublicIp))
	if addr.InstanceId != nil {
		out += fmt.Sprintf(", instance-id: %s", *addr.InstanceId)
	}
	return out
}

// TODO add EBS volume to the VM and attach it to the instance

// TODO ADD A GLOBAL FUNTION THAT WILL HAVE THE ALL OUTPUTS

// Sequence of steps to create a VM
// 1. Create  VPC										DONE
// 2. Create  Subnet									DONE TODO ADD MORE PARAMETERS
// 3. Create  Internet Gateway							DONE  TESTING PENDING
// 4. Create  Route Table								DONE  TESTING PENDING
// 5. Create  Firewall aka Security Group in AWS		DONE  TESTING PENDING
// 6. Create Load Balancer								DONE  TESTING PENDING
// 7. Create Public IP									DONE  TESTING PENDING
// 8. OS IAMGE											TODO  TESTING PENDING
// 9. Generate SSH Key
// 10. Create VM

// TODO Refactor all the code same as various providor

func (obj *AwsProvider) DelVM(factory resources.StorageFactory, i int) error {
	//TODO implement me
	fmt.Println("AWS Del VM")
	return nil
}

var NICID string

func (obj *AwsProvider) CreateNetworkInterface(ctx context.Context, storage resources.StorageFactory, resName string, index int, role string, inistanceid string) (*ec2.CreateNetworkInterfaceOutput, error) {

	// to create networkinterface we need subnetid, securitygroup, availabilityzone, availabilityzoneid, osimage, sshkey, instanceprofile
	// ipv6addresscount, ipv6pool
	interfaceparameter := &ec2.CreateNetworkInterfaceInput{
		Description: aws.String("network interface"),
		Groups: []string{
			awsCloudState.SecurityGroupID,
		},
		SubnetId: aws.String(awsCloudState.SubnetID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("network-interface"),
				Tags: []types.Tag{
					{
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},
	}

	vniclient := obj.ec2Client()
	nicresponse, err := vniclient.CreateNetworkInterface(context.Background(), interfaceparameter)
	if err != nil {
		log.Println(err)
	}

	NICID = *nicresponse.NetworkInterface.NetworkInterfaceId
	storage.Logger().Success("[aws] created the network interface ", *nicresponse.NetworkInterface.NetworkInterfaceId)

	publicip := &ec2.AllocateAddressInput{
		Domain: types.DomainType("vpc"),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("elastic-ip"),
				Tags: []types.Tag{
					{
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},
	}

	publicipresponse, err := vniclient.AllocateAddress(context.Background(), publicip)
	if err != nil {
		log.Println(err)
	}

	// now we need to associate the public ip to the network interface
	_, err = vniclient.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		AllocationId:       publicipresponse.AllocationId,
		InstanceId:         aws.String(inistanceid),
		NetworkInterfaceId: nicresponse.NetworkInterface.NetworkInterfaceId,
	})
	if err != nil {
		log.Println(err)
	}

	time.Sleep(10 * time.Second)
	attachresp, err := vniclient.AttachNetworkInterface(context.Background(), &ec2.AttachNetworkInterfaceInput{
		DeviceIndex:        aws.Int32(0),
		InstanceId:         aws.String(inistanceid),
		NetworkInterfaceId: nicresponse.NetworkInterface.NetworkInterfaceId,
		// SkipSourceDestCheck: aws.Bool(true),
	})
	if err != nil {
		log.Println(err)
	}

	storage.Logger().Success("[aws] attached the network interface ", *nicresponse.NetworkInterface.NetworkInterfaceId)
	storage.Logger().Success("[aws] attached the public ip ", *publicipresponse.PublicIp)
	storage.Logger().Success("[aws] attached the network interface to the instance ", *attachresp.AttachmentId)

	return nicresponse, nil

}

func (obj *AwsProvider) NewVM(storage resources.StorageFactory, indexNo int) error {
	if obj.metadata.role == utils.ROLE_DS && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}

	ec2Client := obj.ec2Client()
	obj.metadata.role = "testing"
	err := obj.NewFirewall(storage)
	if err != nil {
		log.Println(err)
	}

	parameter := &ec2.RunInstancesInput{
		// use awslinux image id ---->   ami-0e306788ff2473ccb
		ImageId:      aws.String("ami-0e306788ff2473ccb"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String("ksctl"),
		SubnetId:     aws.String(awsCloudState.SubnetID),
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
						Key:   aws.String(obj.metadata.resName),
						Value: aws.String("value"),
					},
				},
			},
		},

		// EbsOptimized: aws.Bool(true),
		// BlockDeviceMappings: []types.BlockDeviceMapping{
		// 	{
		// 		DeviceName: aws.String("/dev/sda1"),
		// 		Ebs: &types.EbsBlockDevice{
		// 			DeleteOnTermination: aws.Bool(true),
		// 			VolumeSize:          aws.Int32(8),
		// 			VolumeType:          types.VolumeType("gp2"),
		// 		},
		// 	},
		// },
	}

	instanceop, err := ec2Client.RunInstances(context.Background(), parameter)
	// id := aws.String(*instanceop.Instances[0].InstanceId)
	if err != nil {
		fmt.Println(err)
		panic("Error creating EC2 instance: " + err.Error())
	}

	_, err = obj.CreateNetworkInterface(context.TODO(), storage, obj.metadata.resName, indexNo, obj.metadata.role, *instanceop.Instances[0].InstanceId)
	if err != nil {
		panic("Error creating network interface: " + err.Error())
	}

	// obj.AssignPublicIP(*instanceop.Instances[0].InstanceId, *publicipresponse.PublicIp)

	// if err != nil {
	// 	log.Println(err)
	// }

	storage.Logger().Success("[aws] created the instance ", *instanceop.Instances[0].InstanceId)
	time.Sleep(300 * time.Second)
	return nil
}

func (obj *AwsProvider) DeleteNetworkInterface(ctx context.Context, storage resources.StorageFactory, index int, role string) error {
	return nil
}

func (obj *AwsProvider) DeletePublicIP(ctx context.Context, storage resources.StorageFactory, index int, role string) error {
	return nil
}

func (obj *AwsProvider) CreatePublicIP(ctx context.Context, storage resources.StorageFactory, publicIPName string, index int, role string, instancid string) (*ec2.ReleaseAddressOutput, error) {

	ec2Client := obj.ec2Client()

	allocRes, err := ec2Client.AllocateAddress(ctx, &ec2.AllocateAddressInput{
		Domain: types.DomainTypeVpc,
	})
	if err != nil {
		return nil, err
	}

	_, err = ec2Client.AssociateAddress(ctx, &ec2.AssociateAddressInput{
		AllocationId: allocRes.AllocationId,
		InstanceId:   aws.String(instancid),
	})
	if err != nil {
		return nil, err
	}

	storage.Logger().Success("[aws] created the public IP ", *allocRes.PublicIp)
	return nil, nil
}

func (obj *AwsProvider) DeleteDisk(ctx context.Context, storage resources.StorageFactory, index int, role string) error {

	return nil
}
