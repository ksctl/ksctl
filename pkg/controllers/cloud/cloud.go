package cloud

import (
	"context"
	"fmt"
	"sync"
	"time"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	awsPkg "github.com/ksctl/ksctl/internal/cloudproviders/aws"
	azurePkg "github.com/ksctl/ksctl/internal/cloudproviders/azure"
	civoPkg "github.com/ksctl/ksctl/internal/cloudproviders/civo"
	localPkg "github.com/ksctl/ksctl/internal/cloudproviders/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	log           types.LoggerFactory
	controllerCtx context.Context
)

func InitLogger(ctx context.Context, _log types.LoggerFactory) {
	log = _log
	controllerCtx = ctx
}

func InitCloud(client *types.KsctlClient, state *storageTypes.StorageDocument, operation consts.KsctlOperation, fakeClient bool) error {

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		if !fakeClient {
			client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, state, civoPkg.ProvideClient)
		} else {
			client.Cloud, err = civoPkg.NewClient(controllerCtx, client.Metadata, log, state, civoPkg.ProvideMockClient)
		}

		if err != nil {
			return err
		}
	case consts.CloudAzure:
		if !fakeClient {
			client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, state, azurePkg.ProvideClient)
		} else {
			client.Cloud, err = azurePkg.NewClient(controllerCtx, client.Metadata, log, state, azurePkg.ProvideMockClient)
		}

		if err != nil {
			return err
		}
	case consts.CloudAws:
		if !fakeClient {
			client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, state, awsPkg.ProvideClient)
		} else {
			client.Cloud, err = awsPkg.NewClient(controllerCtx, client.Metadata, log, state, awsPkg.ProvideMockClient)
		}

		if err != nil {
			return err
		}

	case consts.CloudLocal:
		if !fakeClient {
			client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, state, localPkg.ProvideClient)
		} else {
			client.Cloud, err = localPkg.NewClient(controllerCtx, client.Metadata, log, state, localPkg.ProvideMockClient)
		}

		if err != nil {
			return err
		}
	default:
		return log.NewError(controllerCtx, "invalid cloud provider")
	}
	// call the init state for cloud providers
	err = client.Cloud.InitState(client.Storage, operation)
	if err != nil {
		return err
	}
	return nil
}

func pauseOperation(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

func DeleteHACluster(client *types.KsctlClient) error {

	var err error

	noCP, err := client.Cloud.NoOfControlPlane(client.Metadata.NoCP, false)
	if err != nil {
		return err
	}

	noWP, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, false)
	if err != nil {
		return err
	}

	noDS, err := client.Cloud.NoOfDataStore(client.Metadata.NoDS, false)
	if err != nil {
		return err
	}

	//////
	wg := &sync.WaitGroup{}
	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, noDS)
	errChanCP := make(chan error, noCP)
	errChanWP := make(chan error, noWP)

	wg.Add(1 + noDS + noCP + noWP)
	//////
	for no := 0; no < noWP; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(consts.RoleWp).DelVM(client.Storage, no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	for no := 0; no < noCP; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(consts.RoleCp).DelVM(client.Storage, no)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}
	for no := 0; no < noDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(consts.RoleDs).DelVM(client.Storage, no)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}

	go func() {
		defer wg.Done()

		err := client.Cloud.Role(consts.RoleLb).DelVM(client.Storage, 0)
		if err != nil {
			errChanLB <- err
		}
	}()

	////////
	wg.Wait()
	close(errChanDS)
	close(errChanLB)
	close(errChanCP)
	close(errChanWP)

	for err := range errChanLB {
		if err != nil {
			return err
		}
	}
	for err := range errChanDS {
		if err != nil {
			return err
		}
	}
	for err := range errChanCP {
		if err != nil {
			return err
		}
	}
	for err := range errChanWP {
		if err != nil {
			return err
		}
	}

	pauseOperation(20) // NOTE: experimental time to wait for generic cloud to update its state

	err = client.Cloud.Role(consts.RoleDs).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(consts.RoleCp).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(consts.RoleWp).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(consts.RoleLb).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	// MIssing kubeconfig unset printing

	err = client.Cloud.DelSSHKeyPair(client.Storage)
	if err != nil {
		return err
	}

	// NOTE: last one to delete is network
	err = client.Cloud.DelNetwork(client.Storage)
	if err != nil {
		return err
	}

	return nil
}

// AddWorkerNodes the user provides the desired no of workerplane not the no of workerplanes to be added
func AddWorkerNodes(client *types.KsctlClient) (int, error) {

	var err error
	currWP, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, false)
	if err != nil {
		return -1, err
	}

	_, err = client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, true)
	if err != nil {
		return -1, err
	}
	wg := &sync.WaitGroup{}

	errChanWP := make(chan error, client.Metadata.NoWP-currWP)

	for no := currWP; no < client.Metadata.NoWP; no++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", client.Metadata.ClusterName, no)).
				Role(consts.RoleWp).
				VMType(client.Metadata.WorkerPlaneNodeType).
				Visibility(true).
				NewVM(client.Storage, no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChanWP)

	for err := range errChanWP {
		if err != nil {
			return -1, err
		}
	}

	// workerplane created
	return currWP, nil
}

// DelWorkerNodes uses the noWP as the desired count of workerplane which is desired
func DelWorkerNodes(client *types.KsctlClient) ([]string, error) {

	hostnames := client.Cloud.GetHostNameAllWorkerNode()

	if hostnames == nil {
		return nil, log.NewError(controllerCtx, "hostname is empty")
	}

	currLen := len(hostnames)
	desiredLen := client.Metadata.NoWP
	hostnames = hostnames[desiredLen:]

	if desiredLen < 0 || desiredLen > currLen {
		return nil, log.NewError(controllerCtx, "not a valid count of wp for down scaling")
	}

	wg := &sync.WaitGroup{}
	errChanWP := make(chan error, currLen-desiredLen)

	for no := desiredLen; no < currLen; no++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(consts.RoleWp).DelVM(client.Storage, no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChanWP)

	for err := range errChanWP {
		if err != nil {
			return nil, err
		}
	}

	_, err := client.Cloud.NoOfWorkerPlane(client.Storage, desiredLen, true)
	if err != nil {
		return nil, err
	}

	return hostnames, nil

}

func CreateHACluster(client *types.KsctlClient) error {
	var err error
	err = client.Cloud.Name(client.Metadata.ClusterName + "-net").NewNetwork(client.Storage)
	if err != nil {
		return err
	}

	if _, err := client.Cloud.NoOfControlPlane(client.Metadata.NoCP, true); err != nil {
		return err
	}

	if _, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, true); err != nil {
		return err
	}

	if _, err := client.Cloud.NoOfDataStore(client.Metadata.NoDS, true); err != nil {
		return err
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-ssh").CreateUploadSSHKeyPair(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-lb").
		Role(consts.RoleLb).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-db").
		Role(consts.RoleDs).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-cp").
		Role(consts.RoleCp).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-wp").
		Role(consts.RoleWp).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	//////
	wg := &sync.WaitGroup{}
	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, client.Metadata.NoDS)
	errChanCP := make(chan error, client.Metadata.NoCP)
	errChanWP := make(chan error, client.Metadata.NoWP)

	wg.Add(1 + client.Metadata.NoCP + client.Metadata.NoDS + client.Metadata.NoWP)
	//////
	go func() {
		defer wg.Done()

		err := client.Cloud.Name(client.Metadata.ClusterName+"-vm-lb").
			Role(consts.RoleLb).
			VMType(client.Metadata.LoadBalancerNodeType).
			Visibility(true).
			NewVM(client.Storage, 0)
		if err != nil {
			errChanLB <- err
		}
	}()

	for no := 0; no < client.Metadata.NoDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-db-%d", client.Metadata.ClusterName, no)).
				Role(consts.RoleDs).
				VMType(client.Metadata.DataStoreNodeType).
				Visibility(true).
				NewVM(client.Storage, no)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}
	for no := 0; no < client.Metadata.NoCP; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-cp-%d", client.Metadata.ClusterName, no)).
				Role(consts.RoleCp).
				VMType(client.Metadata.ControlPlaneNodeType).
				Visibility(true).
				NewVM(client.Storage, no)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}

	for no := 0; no < client.Metadata.NoWP; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", client.Metadata.ClusterName, no)).
				Role(consts.RoleWp).
				VMType(client.Metadata.WorkerPlaneNodeType).
				Visibility(true).
				NewVM(client.Storage, no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}

	////////
	wg.Wait()
	close(errChanDS)
	close(errChanLB)
	close(errChanCP)
	close(errChanWP)

	for err := range errChanLB {
		if err != nil {
			return err
		}
	}
	for err := range errChanDS {
		if err != nil {
			return err
		}
	}
	for err := range errChanCP {
		if err != nil {
			return err
		}
	}
	for err := range errChanWP {
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateManagedCluster(client *types.KsctlClient) (bool, bool, error) {

	if client.Metadata.Provider != consts.CloudLocal {
		if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(client.Storage); err != nil {
			return false, false, err
		}
	}

	managedClient := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed")

	if client.Metadata.Provider != consts.CloudLocal {
		managedClient = managedClient.VMType(client.Metadata.ManagedNodeType)
	}

	externalApps := managedClient.Application(client.Metadata.Applications)

	externalCNI := managedClient.CNI(client.Metadata.CNIPlugin)

	managedClient = managedClient.ManagedK8sVersion(client.Metadata.K8sVersion)

	if managedClient == nil {
		return externalApps, externalCNI, log.NewError(controllerCtx, "invalid k8s version")
	}

	if err := managedClient.NewManagedCluster(client.Storage, client.Metadata.NoMP); err != nil {
		return externalApps, externalCNI, err
	}
	return externalApps, externalCNI, nil
}

func DeleteManagedCluster(client *types.KsctlClient) error {

	if err := client.Cloud.DelManagedCluster(client.Storage); err != nil {
		return err
	}

	if client.Metadata.Provider != consts.CloudLocal {
		if err := client.Cloud.DelNetwork(client.Storage); err != nil {
			return err
		}
	}
	return nil
}
