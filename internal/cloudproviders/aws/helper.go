package aws

import (
	"encoding/json"
	"fmt"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func generatePath(flag consts.KsctlUtilsConsts, clusterType consts.KsctlClusterType, path ...string) string {
	return utils.GetPath(flag, consts.CloudAws, clusterType, path...)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func saveStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(consts.UtilClusterPath, consts.CloudAws, clusterType, clusterDirName, STATE_FILE_NAME)
	rawState, err := convertStateToBytes(*awsCloudState)
	if err != nil {
		return err
	}

	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func validationOfArguments(obj *AwsProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	if err := utils.IsValidName(obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidRegion(obj *AwsProvider, reg string) error {

	ec2client := obj.ec2Client()

	validReg, err := obj.client.ListLocations(ec2client)
	if err != nil {
		return err
	}
	if validReg == nil {
		return fmt.Errorf("no region found")
	}

	return nil
}

// we need to check vm soxe but aws use consts and we have string
// so will check if the string is in the consts

func isValidVMSize(obj *AwsProvider, size string) error {
	fmt.Println(size)
	fmt.Println("......................................................")
	validSize, err := obj.client.ListVMTypes(obj.ec2Client())
	if err != nil {
		return err
	}

	for _, valid := range validSize.InstanceTypes {
		constAsString := string(valid.InstanceType)
		if constAsString == size {
			return nil
		}
	}

	return fmt.Errorf("INVALID VM SIZE\nValid options %v\n", validSize)
}
