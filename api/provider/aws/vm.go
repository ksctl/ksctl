package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubesimplify/ksctl/api/resources"
	// "github.com/kubesimplify/ksctl/api/provider/aws"
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
	fmt.Print("VPC Created Successfully: ")
	fmt.Println(*vpc.Vpc.VpcId)

}

func (obj *AwsProvider) NewVM(storage resources.StorageFactory, indexNo int) error {

	return nil
}
