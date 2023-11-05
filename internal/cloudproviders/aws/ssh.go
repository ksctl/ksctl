package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {

	name := obj.metadata.resName
	obj.mxName.Unlock()

	if len(awsCloudState.SSHKeyName) != 0 {
		storage.Logger().Success("[skip] ssh key already created", awsCloudState.SSHKeyName)
		return nil
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, CloudAws, clusterDirName)
	if err != nil {
		return err
	}

	parameter := ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: []byte(keyPairToUpload),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(name),
					},
				},
			},
		},
	}

	_, err = obj.ec2Client().ImportKeyPair(context.Background(), &parameter)

	awsCloudState.SSHKeyName = name
	awsCloudState.SSHUser = "ubuntu"
	awsCloudState.SSHPrivateKeyLoc = utils.GetPath(UtilSSHPath, CloudAws, clusterType, clusterDirName)

	if err := saveStateHelper(storage); err != nil {
		return err
	}
	storage.Logger().Success("[aws] created the ssh key pair", awsCloudState.SSHKeyName)

	return nil

}
