package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	//TODO implement me

	obj.mxName.Unlock()

	fmt.Println("AWS Create Upload SSH Key Pair")
	keypairinput := &ec2.CreateKeyPairInput{
		KeyName: aws.String("testkeypair"),
		KeyType: types.KeyTypeRsa,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("testkeypair"),
					},
				},
			},
		},
	}

	response, err := obj.ec2Client().CreateKeyPair(context.Background(), keypairinput)
	if err != nil {
		return err
	}
	
	file, err := os.Create("testkeypair.pem")
	if err != nil {
		return err
	}

	_, err = file.WriteString(*response.KeyMaterial)
	if err != nil {
		return err
	}
	file.Close()



	return nil

}
