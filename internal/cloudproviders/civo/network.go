package civo

import (
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

// NewNetwork implements types.CloudFactory.
func (obj *CivoProvider) NewNetwork(storage types.StorageFactory) error {
	name := <-obj.chResName

	log.Debug(civoCtx, "Printing", "Name", name)

	// check if the networkID already exist
	if len(mainStateDocument.CloudInfra.Civo.NetworkID) != 0 {
		log.Print(civoCtx, "skipped network creation found", "networkID", mainStateDocument.CloudInfra.Civo.NetworkID)
		return nil
	}

	res, err := obj.client.CreateNetwork(name)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Civo.NetworkID = res.ID

	net, err := obj.client.GetNetwork(res.ID)
	if err != nil {
		return err
	}
	mainStateDocument.CloudInfra.Civo.NetworkCIDR = net.CIDR

	log.Debug(civoCtx, "Printing", "networkID", res.ID)
	log.Debug(civoCtx, "Printing", "networkCIDR", net.CIDR)
	log.Success(civoCtx, "Created network", "name", name)

	return storage.Write(mainStateDocument)
}

// DelNetwork implements types.CloudFactory.
func (obj *CivoProvider) DelNetwork(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Civo.NetworkID) == 0 {
		log.Print(civoCtx, "skipped network already deleted")
	} else {
		netID := mainStateDocument.CloudInfra.Civo.NetworkID

		currRetryCounter := consts.KsctlCounterConsts(0)
		for currRetryCounter < consts.CounterMaxWatchRetryCount {
			var err error
			_, err = obj.client.DeleteNetwork(mainStateDocument.CloudInfra.Civo.NetworkID)
			if err != nil {
				currRetryCounter++
				log.Warn(civoCtx, "retrying", "err", err)
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == consts.CounterMaxWatchRetryCount {
			return log.NewError(civoCtx, "failed to delete network timeout")
		}

		mainStateDocument.CloudInfra.Civo.NetworkID = ""
		if err := storage.Write(mainStateDocument); err != nil {
			return err
		}
		log.Success(civoCtx, "Deleted network", "networkID", netID)
	}

	return storage.DeleteCluster()
}
