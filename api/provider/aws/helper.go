package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

func GetInputCredential(storage resources.StorageFactory) error {

	storage.Logger().Print("Enter your ACCESS KEY")
	accessKey, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	storage.Logger().Print("Enter your SECRET KEY")
	secretKey, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	apiStore := Credential{
		AccessKeyID: accessKey,
		Secret:      secretKey,
	}

	if err := utils.SaveCred(storage, apiStore, utils.CLOUD_AWS); err != nil {
		return err
	}

	return nil
}

func generatePath(flag int, path ...string) string {
	return utils.GetPath(flag, utils.CLOUD_AWS, path...)
}

/*
 TODO - USE v2 SDK
func (obj AwsProvider) CreateKeyPair() {
	ec2client := obj.ec2Client()

	ec2client.CreateKeyPair(context.Background(), &ec2.CreateKeyPairInput{
		KeyName:   aws.String(obj.ClusterName),
		KeyType:   types.KeyType("rsa"),
		KeyFormat: types.KeyFormat("pem"),
	})

	awsCloudState.SSHKeyName = obj.ClusterName
	// download the keypair
	// store the keypair in the storage

	ec2client.ImportKeyPair(context.Background(), &ec2.ImportKeyPairInput{
		KeyName: aws.String(obj.ClusterName),
		// TODO remember to change the public key material into the actual key
		PublicKeyMaterial: []byte(""),
	})
}

func saveStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AWS, clusterType, clusterDirName, STATE_FILE_NAME)
	rawState, err := convertStateToBytes(*awsCloudState)
	if err != nil {
		return err
	}
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

*/
