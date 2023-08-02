package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// NewNetwork implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewNetwork(state resources.StateManagementInfrastructure) error {
	res, err := civoClient.NewNetwork(obj.Metadata.ResName)
	if err != nil {
		return err
	}
	civoCloudState.NetworkIDs.NetworkID = res.ID
	fmt.Printf("[civo] Created %s network\n", obj.Metadata.ResName)
	path := generatePath(utils.CLUSTER_PATH, "config.json")
	rawState, err := convertStateToBytes(*civoCloudState)
	if err != nil {
		return err
	}
	return state.Path(path).Save(rawState)
}

// DelNetwork implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelNetwork(state resources.StateManagementInfrastructure) error {
	// state.Get()
	_, err := civoClient.DeleteNetwork(civoCloudState.NetworkIDs.NetworkID)
	if err != nil {
		return err
	}
	fmt.Printf("[civo] Deleted %s network\n", civoCloudState.NetworkIDs.NetworkID)
	// state.Save()
	return nil
}
