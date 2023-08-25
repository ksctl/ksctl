package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubesimplify/ksctl/api/provider/aws"
	"log"
)

func (obj *AwsProvider) ec2Client() *ec2.EC2 {
	ec2client := ec2.New(obj.Session, &aws.Config{
		Region: aws.String(obj.Region),
	})

	//TODO ADD ERROR HANDLING
	return ec2client
}

func (obj *AwsProvider) VpcClient () *ec2.DescribeVpcsOutput {

	vpcClient := 

	return nil
}