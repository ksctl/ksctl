package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/fatih/color"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

var (
	awsCloudState *StateConfiguration
	GatewayID     string
	RouteTableID  string
	VPCID         string
	SUBNETID      []string

	// clusterDirName string
	// clusterType    string // it stores the ha or managed
	// ctx context.Context
)

func (obj *AwsProvider) ec2Client() *ec2.Client {
	ec2client := ec2.NewFromConfig(*obj.Session)
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

func (obj *AwsProvider) CreateSubnet() {

	ec2Client := obj.ec2Client()

	fmt.Print(color.BlueString("Creating Subnet...."))
	subnetClient := ec2.CreateSubnetInput{
		CidrBlock: aws.String("172.31.16.0/20"),
		VpcId:     aws.String(VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("subnet"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-subnet"),
					},
				},
			},
		},
		AvailabilityZone: aws.String("ap-south-1a"),
		// TODO: Add the following parameters
		// AvailabilityZoneId: aws.String(obj.AvailabilityZoneID),
	}

	subnet, err := ec2Client.CreateSubnet(context.TODO(), &subnetClient)
	if err != nil {
		log.Println(err)
	}
	SUBNETID = append(SUBNETID, *subnet.Subnet.SubnetId)

	///////////////////////

	fmt.Print(color.BlueString("Creating Subnet...."))
	subnetClient = ec2.CreateSubnetInput{
		CidrBlock: aws.String("172.31.0.0/20"),
		VpcId:     aws.String(VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("subnet"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-subnet"),
					},
				},
			},
		},
		// TODO: Add the following parameters
		// AvailabilityZoneId: aws.String(obj.AvailabilityZoneID),

		AvailabilityZone: aws.String("ap-south-1b"),
	}

	subnet, err = ec2Client.CreateSubnet(context.TODO(), &subnetClient)
	if err != nil {
		log.Println(err)
	}
	SUBNETID = append(SUBNETID, *subnet.Subnet.SubnetId)

	///////////////////////

	fmt.Print(color.BlueString("Creating Subnet...."))
	subnetClient = ec2.CreateSubnetInput{
		CidrBlock: aws.String("172.31.32.0/20"),
		VpcId:     aws.String(VPCID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("subnet"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-subnet"),
					},
				},
			},
		},
		// TODO: Add the following parameters
		// AvailabilityZoneId: aws.String(obj.AvailabilityZoneID),

		AvailabilityZone: aws.String("ap-south-1c"),
	}

	subnet, err = ec2Client.CreateSubnet(context.TODO(), &subnetClient)
	if err != nil {
		log.Println(err)
	}

	SUBNETID = append(SUBNETID, *subnet.Subnet.SubnetId)
	fmt.Print("Subnet Created Successfully: ")
	fmt.Println(*subnet.Subnet.SubnetId)
	// awsCloudState.SubnetID = *subnet.Subnet.SubnetId
	// fmt.Println("fromstate", awsCloudState.SubnetID)
}

func (obj *AwsProvider) CreateInternetGateway() {

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
	elbv2Client := elasticloadbalancingv2.NewFromConfig(*obj.Session)

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

	client.RegisterTargets(context.TODO(), &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(*ARNV),
		Targets: []elb_types.TargetDescription{
			{
				Id: aws.String(""),
				// TODO: Add the more parameters
			},
		},
	})

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
	switch obj.Metadata.Role {
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
						Key:   aws.String(obj.Metadata.ResName),
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
		fmt.Printf("No elastic IPs for %s region\n", obj.Region)
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
						Key:   aws.String(obj.Metadata.ResName),
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

func (obj *AwsProvider) AssignPublicIP(instanceid string) error {
	client := obj.ec2Client()
	_, err := client.AssociateAddress(context.Background(), &ec2.AssociateAddressInput{
		InstanceId: aws.String(instanceid),
		PublicIp:   aws.String(""),
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
func (obj *AwsProvider) NewVM(storage resources.StorageFactory, indexNo int) error {
	if obj.Metadata.Role == utils.ROLE_DS && indexNo > 0 {
		storage.Logger().Note("[skip] currently multiple datastore not supported")
		return nil
	}

	ec2Client := obj.ec2Client()

	VM, err := ec2Client.RunInstances(context.TODO(), &ec2.RunInstancesInput{
		/*
			TODO: Add the following parameters
			vmname
			vpcid
			subnetid
			securitygroup
			availabilityzone
			availabilityzoneid
			osimage
			sshkey
			instanceprofile t2.micro
		*/
		ImageId:      aws.String("ami-e7527ed7"),
		InstanceType: types.InstanceType("t2.micro"),
		KeyName:      aws.String(awsCloudState.SSHKeyName),

		// TODO figure out add the main parameters or not
		// MaintenanceOptions: ,

		MaxCount: aws.Int32(1), //this is the number of instances you want to create
		MinCount: aws.Int32(1), //this is the number of instances you want to create

		// Monitoring: types.RunInstancesMonitoringEnabled{
		// 	Enabled: aws.Bool(true),
		// },

		PrivateIpAddress: aws.String(""),

		EnablePrimaryIpv6: aws.Bool(true),
		SecurityGroupIds: []string{
			awsCloudState.SecurityGroupID,
		},
		SubnetId: aws.String(awsCloudState.SubnetID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceType("instance"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.Metadata.ResName),
					},
				},
			},
		},

		// capacity reservation is used to reserve the capacity for the VM aka the storage
		// CapacityReservationSpecification: &ec2.CapacityReservationSpecification{
		// 	CapacityReservationPreference: aws.String("open"),
		// 	CapacityReservationTarget: &ec2.CapacityReservationTarget{
		// 		CapacityReservationId: aws.String("string"),
		// 		CapacityReservationResourceGroupArn: aws.String("string"),
		// 	},
		// },

		// add disk size
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println("VM Created Successfully....")
	fmt.Println(VM.Instances[0].InstanceId)
	fmt.Println(VM.Instances[0].SecurityGroups)
	fmt.Println(VM.Instances[0].PublicIpAddress)
	return nil
}
