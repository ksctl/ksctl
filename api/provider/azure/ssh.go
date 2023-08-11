package azure

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// CreateUploadSSHKeyPair implements resources.CloudFactory.
func (cloud *AzureProvider) CreateUploadSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudFactory.
func (*AzureProvider) DelSSHKeyPair(state resources.StorageFactory) error {
	panic("unimplemented")
}

func (obj *AzureProvider) DeleteSSHKeyPair(ctx context.Context, storage resources.StorageFactory) error {
	sshClient, err := armcompute.NewSSHPublicKeysClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return err
	}
	_, err = sshClient.Delete(ctx, azureCloudState.ResourceGroupName, azureCloudState.SSHKeyName, nil)
	if err != nil {
		return err
	}

	storage.Logger().Success("Deleted the ssh", azureCloudState.SSHKeyName)
	return nil
}

func (obj *AzureProvider) UploadSSHKey(ctx context.Context, storage resources.StorageFactory) (err error) {
	// NOTE: use resName
	sshClient, err := armcompute.NewSSHPublicKeysClient(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		return
	}
	path := utils.GetPath(utils.CLUSTER_PATH, "azure", "ha", obj.ClusterName+" "+azureCloudState.ResourceGroupName+" "+obj.Region)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}

	keyPairToUpload, err := utils.CreateSSHKeyPair(storage, "azure", obj.ClusterName+" "+azureCloudState.ResourceGroupName+" "+obj.Region)
	if err != nil {
		return
	}

	_, err = sshClient.Create(ctx, azureCloudState.ResourceGroupName, obj.ClusterName+"-ssh", armcompute.SSHPublicKeyResource{
		Location:   to.Ptr(obj.Region),
		Properties: &armcompute.SSHPublicKeyResourceProperties{PublicKey: to.Ptr(keyPairToUpload)},
	}, nil)
	azureCloudState.SSHKeyName = obj.ClusterName + "-ssh"

	// ------- Setting the ssh configs only the public ips used will change
	azureCloudState.SSHUser = "azureuser"
	azureCloudState.SSHPrivateKeyLoc = utils.GetPath(utils.SSH_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName)
	// ------

	return
}
