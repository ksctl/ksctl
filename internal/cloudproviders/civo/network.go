package civo

import (
	"fmt"
	"time"

	"github.com/kubesimplify/ksctl/pkg/resources"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// NewNetwork implements resources.CloudFactory.
func (obj *CivoProvider) NewNetwork(storage resources.StorageFactory) error {
	name := obj.metadata.resName
	obj.mxName.Unlock()

	// check if the networkID already exist
	if len(civoCloudState.NetworkIDs.NetworkID) != 0 {
		storage.Logger().Success("[skip] network creation found", civoCloudState.NetworkIDs.NetworkID)
		return nil
	}

	res, err := obj.client.CreateNetwork(name)
	if err != nil {
		return err
	}
	civoCloudState.NetworkIDs.NetworkID = res.ID
	storage.Logger().Success("[civo] Created network", name)

	// NOTE: as network creation marks first resource we should create the directoy
	// when its success

	if err := storage.Path(generatePath(UtilClusterPath, clusterType, clusterDirName)).
		Permission(FILE_PERM_CLUSTER_DIR).CreateDir(); err != nil {
		return err
	}

	path := generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)

	return saveStateHelper(storage, path)
}

// DelNetwork implements resources.CloudFactory.
func (obj *CivoProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(civoCloudState.NetworkIDs.NetworkID) == 0 {
		storage.Logger().Success("[skip] network already deleted")
	} else {

		currRetryCounter := KsctlCounterConsts(0)
		for currRetryCounter < CounterMaxWatchRetryCount {
			var err error
			_, err = obj.client.DeleteNetwork(civoCloudState.NetworkIDs.NetworkID)
			if err != nil {
				currRetryCounter++
				storage.Logger().Warn(fmt.Sprintln("RETRYING", err))
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == CounterMaxWatchRetryCount {
			return fmt.Errorf("[civo] failed to delete network timeout")
		}

		civoCloudState.NetworkIDs.NetworkID = ""
		if err := saveStateHelper(storage, generatePath(UtilClusterPath, clusterType, clusterDirName, STATE_FILE_NAME)); err != nil {
			return err
		}
		storage.Logger().Success("[civo] Deleted network", civoCloudState.NetworkIDs.NetworkID)
	}
	path := generatePath(UtilClusterPath, clusterType, clusterDirName)
	return storage.Path(path).DeleteDir()
}
