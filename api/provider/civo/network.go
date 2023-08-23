package civo

import (
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"time"
)

// NewNetwork implements resources.CloudFactory.
func (obj *CivoProvider) NewNetwork(storage resources.StorageFactory) error {

	// check if the networkID already exist
	if len(civoCloudState.NetworkIDs.NetworkID) != 0 {
		storage.Logger().Success("[skip] network creation found", civoCloudState.NetworkIDs.NetworkID)
		return nil
	}

	res, err := obj.Client.CreateNetwork(obj.Metadata.ResName)
	if err != nil {
		return err
	}
	civoCloudState.NetworkIDs.NetworkID = res.ID
	storage.Logger().Success("[civo] Created network", obj.Metadata.ResName)

	// NOTE: as network creation marks first resource we should create the directoy
	// when its success

	if err := storage.Path(generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return err
	}

	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)

	return saveStateHelper(storage, path)
}

// DelNetwork implements resources.CloudFactory.
func (obj *CivoProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(civoCloudState.NetworkIDs.NetworkID) == 0 {
		storage.Logger().Success("[skip] network already deleted")
	} else {

		currRetryCounter := 0
		for currRetryCounter < utils.MAX_WATCH_RETRY_COUNT {
			var err error
			_, err = obj.Client.DeleteNetwork(civoCloudState.NetworkIDs.NetworkID)
			if err != nil {
				currRetryCounter++
				storage.Logger().Err(fmt.Sprintln("RETRYING", err))
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == utils.MAX_WATCH_RETRY_COUNT {
			return fmt.Errorf("[civo] failed to delete network timeout")
		}

		civoCloudState.NetworkIDs.NetworkID = ""
		if err := saveStateHelper(storage, generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, STATE_FILE_NAME)); err != nil {
			return err
		}
		storage.Logger().Success("[civo] Deleted network", civoCloudState.NetworkIDs.NetworkID)
	}
	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName)
	return storage.Path(path).DeleteDir()
}
