package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/fatih/color"
	ksctlTypes "github.com/kubesimplify/ksctl/internal/storage/types"
	"github.com/kubesimplify/ksctl/pkg/helpers"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

var ksctlLog resources.LoggerFactory = logger.NewDefaultLogger(0, os.Stdout)

func createSG(client *ec2.Client) (*string, error) {
	sgInput := &ec2.CreateSecurityGroupInput{
		Description: aws.String("testrun"),
		GroupName:   aws.String("test"),
	}

	// Create the security group
	createSGResp, err := client.CreateSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create security group: %w", err)
	}

	// Get the ID of the created security group
	securityGroupID := createSGResp.GroupId

	// Add an inbound rule to allow incoming traffic on port 22 (SSH)
	ingressInput := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    securityGroupID,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(22),
		ToPort:     aws.Int32(22),
		CidrIp:     aws.String("0.0.0.0/0"), // Allow from any IP address
	}

	// Authorize the security group ingress rule
	_, err = client.AuthorizeSecurityGroupIngress(context.TODO(), ingressInput)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize ingress rule: %w", err)
	}
	return securityGroupID, nil
}

func CreateInstance(client *ec2.Client, sgId *string) (*string, error) {
	amiID, err := getLatestAmazonLinux2AMI(client)
	if err != nil {
		log.Fatalf("failed to get AMI ID: %v", err)
		return nil, err
	}
	fmt.Println("amiid:", amiID)

	ksctlState := &ksctlTypes.StorageDocument{}
	if err := helpers.CreateSSHKeyPair(ksctlLog, ksctlState); err != nil {
		return nil, err
	}
	sshPrivatePath := path.Join(os.TempDir(), "demo-ssh-keypair-rsa.pem")

	if err := os.WriteFile(sshPrivatePath, []byte(ksctlState.SSHKeyPair.PrivateKey), 0755); err != nil {
		return nil, err
	}

	parameter := &ec2.ImportKeyPairInput{
		KeyName:           aws.String("demo-ssh-keypair-rsa"),
		PublicKeyMaterial: []byte(ksctlState.SSHKeyPair.PublicKey),
	}

	if _, err := client.ImportKeyPair(context.Background(), parameter); err != nil {
		return nil, err
	}

	instanceInput := &ec2.RunInstancesInput{
		ImageId:          aws.String(amiID),
		InstanceType:     types.InstanceTypeT2Micro,
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		KeyName:          aws.String("demo-ssh-keypair-rsa"),
		SecurityGroupIds: []string{*sgId},
	}

	resp, err := client.RunInstances(context.TODO(), instanceInput)
	if err != nil {
		log.Fatalf("failed to launch instance, %v", err)
		return nil, err
	}

	instanceID := resp.Instances[0].InstanceId
	fmt.Printf("Launched EC2 instance with ID: %s\n", *instanceID)
	ec2RunningWaiter := ec2.NewInstanceRunningWaiter(client, func(irwo *ec2.InstanceRunningWaiterOptions) {
		irwo.MaxDelay = 300 * time.Second
		irwo.MinDelay = 20 * time.Second
	})

	describeEc2Inp := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceID},
	}

	err = ec2RunningWaiter.Wait(context.TODO(), describeEc2Inp, 300*time.Second)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("Your instance is up and running, id=", *instanceID)

	out, err := client.DescribeInstances(context.TODO(), describeEc2Inp)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Check if any instances were found
	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found with ID: %s", *instanceID)
	}

	// Get the public IP address of the instance (assuming it has only one network interface)
	pubIP := out.Reservations[0].Instances[0].PublicIpAddress
	privateIP := out.Reservations[0].Instances[0].PrivateIpAddress
	fmt.Println("PrivateIP=", privateIP)
	color.Green("Do your work using the above information and continue for destruction")
	color.HiYellow("execute the following command to try out ssh execution")

	fmt.Printf("\ngo run ssh_main.go -user=ubuntu -ip=%s -ssh-key=%s\n\n", *pubIP, sshPrivatePath)
	return instanceID, nil
}

func DeleteInstance(client *ec2.Client, instanceID *string) error {

	if _, err := client.DeleteKeyPair(context.Background(), &ec2.DeleteKeyPairInput{KeyName: aws.String("demo-ssh-keypair-rsa")}); err != nil {
		return err
	}

	_, err := client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{InstanceIds: []string{*instanceID}})
	if err != nil {
		log.Fatalf("failed to delete instance, %v", err)
		return err
	}

	ec2TerminatedWaiter := ec2.NewInstanceTerminatedWaiter(client, func(itwo *ec2.InstanceTerminatedWaiterOptions) {
		itwo.MaxDelay = 300 * time.Second
		itwo.MinDelay = 20 * time.Second
	})

	describeEc2Inp := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceID},
	}

	err = ec2TerminatedWaiter.Wait(context.TODO(), describeEc2Inp, 300*time.Second)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func deleteSG(client *ec2.Client, sgId *string) error {

	// Create the security group
	_, err := client.DeleteSecurityGroup(context.TODO(), &ec2.DeleteSecurityGroupInput{GroupId: sgId})
	if err != nil {
		return fmt.Errorf("failed to create security group: %w", err)
	}

	return nil
}

func waiter() {
	fmt.Println("Enter 0 to continue towards deletion")
	ch := 0
	_, err := fmt.Scanf("%d", &ch)
	if err != nil {
		panic(err)
	}
	if ch == 0 {
		return
	} else {
		fmt.Println("Existed...")
		os.Exit(0)
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile("ksctl"), config.WithRegion("ap-south-1"))
	if err != nil {
		log.Fatal(err)
	}
	client := ec2.NewFromConfig(cfg)

	sgId, err := createSG(client)
	if err != nil {
		panic(err)
	}
	instID, err := CreateInstance(client, sgId)
	if err != nil {
		panic(err)
	}

	waiter()

	if err := DeleteInstance(client, instID); err != nil {
		panic(err)
	}

	if err := deleteSG(client, sgId); err != nil {
		panic(err)
	}

	fmt.Println("Delete the Resources")
}
