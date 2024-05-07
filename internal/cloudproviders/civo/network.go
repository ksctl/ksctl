package civo

import (
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// NewNetwork implements resources.CloudFactory.
func (obj *CivoProvider) NewNetwork(storage resources.StorageFactory) error {
	name := <-obj.chResName

	log.Debug("Printing", "Name", name)

	// check if the networkID already exist
	if len(mainStateDocument.CloudInfra.Civo.NetworkID) != 0 {
		log.Print("skipped network creation found", "networkID", mainStateDocument.CloudInfra.Civo.NetworkID)
		return nil
	}

	res, err := obj.client.CreateNetwork(name)
	if err != nil {
		return log.NewError(err.Error())
	}
	mainStateDocument.CloudInfra.Civo.NetworkID = res.ID

	net, err := obj.client.GetNetwork(res.ID)
	if err != nil {
		return log.NewError(err.Error())
	}
	mainStateDocument.CloudInfra.Civo.NetworkCIDR = net.CIDR

	log.Debug("Printing", "networkID", res.ID)
	log.Debug("Printing", "networkCIDR", net.CIDR)
	log.Success("Created network", "name", name)

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

// DelNetwork implements resources.CloudFactory.
func (obj *CivoProvider) DelNetwork(storage resources.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Civo.NetworkID) == 0 {
		log.Print("skipped network already deleted")
	} else {
		netID := mainStateDocument.CloudInfra.Civo.NetworkID

		currRetryCounter := consts.KsctlCounterConsts(0)
		for currRetryCounter < consts.CounterMaxWatchRetryCount {
			var err error
			_, err = obj.client.DeleteNetwork(mainStateDocument.CloudInfra.Civo.NetworkID)
			if err != nil {
				currRetryCounter++
				log.Warn("RETRYING", err)
			} else {
				break
			}
			time.Sleep(5 * time.Second)
		}
		if currRetryCounter == consts.CounterMaxWatchRetryCount {
			return log.NewError("failed to delete network timeout")
		}

		mainStateDocument.CloudInfra.Civo.NetworkID = ""
		if err := storage.Write(mainStateDocument); err != nil {
			return log.NewError(err.Error())
		}
		log.Success("Deleted network", "networkID", netID)
	}

	if err := storage.DeleteCluster(); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
