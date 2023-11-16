package aws

import (
	"encoding/json"
	"fmt"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

func generatePath(flag consts.KsctlUtilsConsts, clusterType consts.KsctlClusterType, path ...string) string {
	return utils.GetPath(flag, consts.CloudAws, clusterType, path...)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
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

func (obj *AwsProvider) GetKubeconfigPath() string {

	return ""
}

func isValidK8sVersion(obj *AwsProvider, ver string) error {
	res, err := obj.client.ListKubernetesVersions()
	if err != nil {
		return log.NewError("failed to finish the request: %v", err)
	}

	log.Debug("Printing", "ListKubernetesVersions", res)

	var vers []string
	for _, version := range res.Values {
		vers = append(vers, *version.Version)
	}
	for _, valver := range vers {
		if valver == ver {
			return nil
		}
	}
	return log.NewError("Invalid k8s version\nValid options: %v\n", vers)
}
