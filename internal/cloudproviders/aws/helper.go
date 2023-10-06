package aws

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/utils"
)

func generatePath(flag int, path ...string) string {
	return utils.GetPath(flag, utils.CLOUD_AWS, path...)
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
		return fmt.Errorf("No region found")
	}

	return nil
}
