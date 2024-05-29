package civo

import (
	"fmt"
	"time"

	"github.com/civo/civogo"
	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func (obj *CivoProvider) NewNetwork(storage types.StorageFactory) error {
	name := <-obj.chResName

	log.Debug(civoCtx, "Printing", "Name", name)

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

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}

	expoBackoff := helpers.NewBackOff(
		10*time.Second,
		2,
		int(consts.CounterMaxWatchRetryCount),
	)
	var netInst *civogo.Network
	_err := expoBackoff.Run(
		civoCtx,
		log,
		func() (err error) {
			netInst, err = obj.client.GetNetwork(res.ID)
			return err
		},
		func() bool {
			return netInst.Status == "Active"
		},
		nil,
		func() error {
			log.Print(civoCtx, "network ready", "name", name)
			return nil
		},
		fmt.Sprintf("Waiting for the network %s to be ready", name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (obj *CivoProvider) DelNetwork(storage types.StorageFactory) error {

	if len(mainStateDocument.CloudInfra.Civo.NetworkID) == 0 {
		log.Print(civoCtx, "skipped network already deleted")
		return nil
	}
	netId := mainStateDocument.CloudInfra.Civo.NetworkID

	expoBackoff := helpers.NewBackOff(
		5*time.Second,
		2,
		int(consts.CounterMaxWatchRetryCount),
	)
	_err := expoBackoff.Run(
		civoCtx,
		log,
		func() (err error) {
			_, err = obj.client.DeleteNetwork(mainStateDocument.CloudInfra.Civo.NetworkID)
			return err
		},
		func() bool {
			return true
		},
		nil,
		func() error {
			mainStateDocument.CloudInfra.Civo.NetworkID = ""
			return storage.Write(mainStateDocument)
		},
		fmt.Sprintf("Waiting for the network %s to be deleted", mainStateDocument.CloudInfra.Civo.NetworkID),
	)
	if _err != nil {
		return _err
	}

	log.Success(civoCtx, "Deleted network", "networkID", netId)

	return storage.DeleteCluster()
}
