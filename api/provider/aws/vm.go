package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	// "github.com/kubesimplify/ksctl/api/provider/aws"
)

var (
	awsCloudState *StateConfiguration

	// clusterDirName string
	// clusterType    string // it stores the ha or managed
	// ctx context.Context
)

func (obj *AwsProvider) ec2Client() *ec2.EC2 {
	ec2client := ec2.New(obj.Session, &aws.Config{
		Region: aws.String(obj.Region),
	})

	//TODO ADD ERROR HANDLING
	fmt.Println("EC2 Client Created Successfully")
	return ec2client
}

func (obj *AwsProvider) vpcClienet() ec2.CreateVpcInput {

	vpcClient := ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
		// Dry run is used to check if the request is valid
		// without actually creating the VPC.
	}
	fmt.Println("VPC Client Created Successfully")
	return vpcClient

}

func (obj *AwsProvider) CreateVPC() {

	vpcClient := obj.vpcClienet()
	ec2Client := obj.ec2Client()

	vpc, err := ec2Client.CreateVpc(&vpcClient)
	if err != nil {
		log.Println(err)
	}
	awsCloudState.VPC = *vpc.Vpc.VpcId
	fmt.Print("VPC Created Successfully: ")
	// fmt.Println(*vpc.Vpc.VpcId)

}

func (obj *AwsProvider) CreateSubnet() {

	ec2Client := obj.ec2Client()

	subnetClient := ec2.CreateSubnetInput{
		CidrBlock: aws.String("10.0.1.0/24"),
		VpcId:     aws.String(awsCloudState.VPC),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("subnet"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-subnet"),
					},
				},
			},
		},

		// TODO: Add the following parameters
		// AvailabilityZone:   aws.String(obj.AvailabilityZone),
		// AvailabilityZoneId: aws.String(obj.AvailabilityZoneID),
		Ipv6Native: aws.Bool(true),
	}

	subnet, err := ec2Client.CreateSubnet(&subnetClient)
	if err != nil {
		log.Println(err)
	}

	fmt.Print("Subnet Created Successfully: ")
	fmt.Println(*subnet.Subnet.SubnetId)
	awsCloudState.SubnetID = *subnet.Subnet.SubnetId
	// fmt.Println("fromstate", awsCloudState.SubnetID)
}

func (obj *AwsProvider) CreateInternetGateway() {

	ec2Client := obj.ec2Client()

	internetGatewayClient := ec2.CreateInternetGatewayInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("internet-gateway"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-ig"),
					},
				},
			},
		},
	}

	internetGateway, err := ec2Client.CreateInternetGateway(&internetGatewayClient)
	if err != nil {
		log.Println(err)
	}

	_, err = ec2Client.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*internetGateway.InternetGateway.InternetGatewayId),
		VpcId:             aws.String(awsCloudState.VPC),
	})

	fmt.Println(*internetGateway.InternetGateway.InternetGatewayId)
	awsCloudState.GatewayID = *internetGateway.InternetGateway.InternetGatewayId
	fmt.Print("Internet Gateway Created Successfully: ")
}

func (obj *AwsProvider) CreateRouteTable() {

	ec2Client := obj.ec2Client()

	routeTableClient := ec2.CreateRouteTableInput{
		VpcId: aws.String(awsCloudState.VPC),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("route-table"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(obj.ClusterName + "-rt"),
					},
				},
			},
		},
	}

	routeTable, err := ec2Client.CreateRouteTable(&routeTableClient)
	if err != nil {
		log.Println(err)
	}

	fmt.Print("Route Table Created Successfully: ")
	fmt.Println(*routeTable.RouteTable.RouteTableId)
	awsCloudState.RouteTableID = *routeTable.RouteTable.RouteTableId

	//  associate route table id with subnet
	_, err = ec2Client.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeTable.RouteTable.RouteTableId),
		SubnetId:     aws.String(awsCloudState.SubnetID),
	})
	fmt.Println("Route Table Associated Successfully....")
	// create route
	_, err = ec2Client.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(awsCloudState.GatewayID),
		RouteTableId:         aws.String(*routeTable.RouteTable.RouteTableId),
	})

	fmt.Println("Route Created Successfully....")

}

/*
	create lb   				DONE
	create target group			DONE
	register target group		DONE
	create listener				DONE
*/

func (obj *AwsProvider) ElbClient() *elbv2.ELBV2 {
	LBCLIENT := elbv2.New(obj.Session, &aws.Config{
		Region: aws.String(obj.Region),
	})
	return LBCLIENT
}

var (
	LBARN *elbv2.CreateLoadBalancerOutput
	ARN   *elbv2.CreateTargetGroupOutput
)

func (obj *AwsProvider) CreateLoadBalancer() (*elbv2.CreateLoadBalancerOutput, error) {
	LBCLIENT := obj.ElbClient()

	LB_ARN, err := LBCLIENT.CreateLoadBalancer(&elbv2.CreateLoadBalancerInput{
		Name: aws.String(obj.ClusterName + "-lb"),

		SecurityGroups: []*string{
			aws.String(awsCloudState.SecurityGroupID),
		},
		Subnets: []*string{
			aws.String(awsCloudState.SubnetID),
		},
		Tags: []*elbv2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(obj.ClusterName + "-lb"),
			},
		},
		Type: aws.String("application"),
		// SubnetMappings: []*elbv2.SubnetMapping{},
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Created Load Balancer Successfully: ", *LB_ARN.LoadBalancers[0].LoadBalancerArn)
	fmt.Println("Creating required resources for Load Balancer....")
	LBARN = LB_ARN
	return LB_ARN, nil

}

func (obj *AwsProvider) CreateTargetGroup() (*elbv2.CreateTargetGroupOutput, error) {
	LBCLIENT := obj.ElbClient()

	obj.CreateLoadBalancer()

	ARN, err := LBCLIENT.CreateTargetGroup(&elbv2.CreateTargetGroupInput{
		Name:       aws.String(obj.ClusterName + "-tg"),
		Protocol:   aws.String("tcp"),
		Port:       aws.Int64(6443), //  port on which the target receives traffic from the load balancer
		VpcId:      aws.String(awsCloudState.VPC),
		TargetType: aws.String("ip"),
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Target Group Created Successfully: ", *ARN.TargetGroups[0].TargetGroupArn)

	return ARN, nil

}

func (obj *AwsProvider) RegisterTargetGroup() {
	client := obj.ElbClient()

	ARN, err := obj.CreateTargetGroup()
	if err != nil {
		log.Println(err)
	}

	ARNV := ARN.TargetGroups[0].TargetGroupArn

	client.RegisterTargets(&elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(*ARNV),
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String("id"),
			},
		},
	})

	fmt.Println("Target Group Registered Successfully: ", *LBARN.LoadBalancers[0].LoadBalancerArn)
}

func (obj *AwsProvider) CreateListener() {
	client := obj.ElbClient()

	ARN, err := obj.CreateTargetGroup()
	if err != nil {
		log.Println(err)
	}

	ARNV := ARN.TargetGroups[0].TargetGroupArn

	client.CreateListener(&elbv2.CreateListenerInput{
		DefaultActions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(*ARNV),
				Type:           aws.String("forward"),
			},
		},
		LoadBalancerArn: aws.String(*LBARN.LoadBalancers[0].LoadBalancerArn),
		Port:            aws.Int64(6443),
		Protocol:        aws.String("TCP"),
	})

	fmt.Println("Listener Created Successfully: ", *LBARN.LoadBalancers[0].LoadBalancerArn)

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
	Ipv4Pool, err := client.CreatePublicIpv4Pool(&ec2.CreatePublicIpv4PoolInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("public--ip-pool"),
				Tags: []*ec2.Tag{
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
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
		},
		AllocationIds: []*string{
			aws.String(*Ipv4Pool.PoolId),
		},
	}
	ipaddress, err := client.DescribeAddresses(parameters)
	if err != nil {
		log.Println(err)
	}
	if ipaddress.Addresses == 0 {
		fmt.Printf("No elastic IPs for %s region\n", obj.Region)
		return err
	}
	fmt.Println("Elastic IPs")
	for _, addr := range ipaddress.Addresses {
		fmt.Println("*", fmtAddress(addr))
	}

	// TODO its just a pool create the public ip
	allocateIp, err := client.AllocateAddress(&ec2.AllocateAddressInput{
		Domain:                aws.String("vpc"),
		CustomerOwnedIpv4Pool: aws.String(*Ipv4Pool.PoolId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("elastic-ip"),
				Tags: []*ec2.Tag{
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
	_, err := client.AssociateAddress(&ec2.AssociateAddressInput{
		InstanceId: aws.String(instanceid),
		PublicIp:   aws.String(""),
	})
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("Public IP %s assigned to %s", "ip", instanceid)

	return nil
}

func fmtAddress(addr *ec2.Address) string {
	out := fmt.Sprintf("IP: %s,  allocation id: %s",
		aws.StringValue(addr.PublicIp), aws.StringValue(addr.AllocationId))
	if addr.InstanceId != nil {
		out += fmt.Sprintf(", instance-id: %s", *addr.InstanceId)
	}
	return out
}

// TODO add EBS volume to the VM and attach it to the instance

// TODO ADD A GLOBAL FUNTION THAT WILL HAVE THE ALL OUTPUTS

func (obj *AwsProvider) randdom() {
	op, err := obj.CreateLoadBalancer()
	if err != nil {
		log.Println(err)
	}

	fmt.Println(op.LoadBalancers[0].LoadBalancerArn)
}

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
	// pubIPName := obj.Metadata.ResName + "-pub"
	// nicName := obj.Metadata.ResName + "-nic"
	// diskName := obj.Metadata.ResName + "-disk"

	ec2Client := obj.ec2Client()

	VM, err := ec2Client.RunInstances(&ec2.RunInstancesInput{
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
		InstanceType: aws.String("t2.micro"),
		KeyName:      aws.String(awsCloudState.SSHKeyName),

		// TODO figure out add the main parameters or not
		// MaintenanceOptions: ,

		MaxCount: aws.Int64(1), //this is the number of instances you want to create
		MinCount: aws.Int64(1), //this is the number of instances you want to create

		Monitoring: &ec2.RunInstancesMonitoringEnabled{
			Enabled: aws.Bool(true),
		},

		PrivateIpAddress: aws.String(""),

		EnablePrimaryIpv6: aws.Bool(true),
		SecurityGroupIds: []*string{
			aws.String(awsCloudState.SecurityGroupID),
		},
		SubnetId: aws.String(awsCloudState.SubnetID),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
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
