package cloud

import (
	"fmt"
	"sync"
	"time"

	awspkg "github.com/kubesimplify/ksctl/internal/cloudproviders/aws"
	azurePkg "github.com/kubesimplify/ksctl/internal/cloudproviders/azure"
	civoPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/civo"
	localPkg "github.com/kubesimplify/ksctl/internal/cloudproviders/local"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

var log resources.LoggerFactory

func HydrateCloud(client *resources.KsctlClient, operation consts.KsctlOperation, fakeClient bool) error {

	log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
	log.SetPackageName("ksctl-cloud")

	var err error
	switch client.Metadata.Provider {
	case consts.CloudCivo:
		if !fakeClient {
			client.Cloud, err = civoPkg.ReturnCivoStruct(client.Metadata, civoPkg.ProvideClient)
		} else {
			client.Cloud, err = civoPkg.ReturnCivoStruct(client.Metadata, civoPkg.ProvideMockCivoClient)
		}

		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudAzure:
		if !fakeClient {
			client.Cloud, err = azurePkg.ReturnAzureStruct(client.Metadata, azurePkg.ProvideClient)
		} else {
			client.Cloud, err = azurePkg.ReturnAzureStruct(client.Metadata, azurePkg.ProvideMockClient)
		}

		if err != nil {
			return log.NewError(err.Error())
		}
	case consts.CloudAws:
		if !fakeClient {
			client.Cloud, err = awspkg.ReturnAwsStruct(client.Metadata, awspkg.ProvideClient)
		} else {
			client.Cloud, err = awspkg.ReturnAwsStruct(client.Metadata, awspkg.ProvideMockClient)
		}
		if err != nil {
			return fmt.Errorf("[cloud] " + err.Error())
		}

	case consts.CloudLocal:
		if !fakeClient {
			client.Cloud, err = localPkg.ReturnLocalStruct(client.Metadata, localPkg.ProvideClient)
		} else {
			client.Cloud, err = localPkg.ReturnLocalStruct(client.Metadata, localPkg.ProvideMockClient)
		}

		if err != nil {
			return log.NewError(err.Error())
		}
	default:
		return log.NewError("invalid cloud provider")
	}
	// call the init state for cloud providers
	err = client.Cloud.InitState(client.Storage, operation)
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func pauseOperation(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

func DeleteHACluster(client *resources.KsctlClient) error {

	var err error

	noCP, err := client.Cloud.NoOfControlPlane(client.Metadata.NoCP, false)
	if err != nil {
		return log.NewError(err.Error())
	}

	noWP, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, false)
	if err != nil {
		return log.NewError(err.Error())
	}

	noDS, err := client.Cloud.NoOfDataStore(client.Metadata.NoDS, false)
	if err != nil {
		return log.NewError(err.Error())
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
			return log.NewError(err.Error())
		}
	}
	for err := range errChanDS {
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	for err := range errChanCP {
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	for err := range errChanWP {
		if err != nil {
			return log.NewError(err.Error())
		}
	}

	pauseOperation(20) // NOTE: experimental time to wait for generic cloud to update its state

	err = client.Cloud.Role(consts.RoleDs).DelFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Role(consts.RoleCp).DelFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Role(consts.RoleWp).DelFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Role(consts.RoleLb).DelFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	// MIssing kubeconfig unset printing

	err = client.Cloud.DelSSHKeyPair(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	// NOTE: last one to delete is network
	err = client.Cloud.DelNetwork(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	return nil
}

// AddWorkerNodes the user provides the desired no of workerplane not the no of workerplanes to be added
func AddWorkerNodes(client *resources.KsctlClient) (int, error) {

	var err error
	currWP, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, false)
	if err != nil {
		return -1, log.NewError(err.Error())
	}

	_, err = client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, true)
	if err != nil {
		return -1, log.NewError(err.Error())
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
			return -1, log.NewError(err.Error())
		}
	}

	// workerplane created
	return currWP, nil
}

// DelWorkerNodes uses the noWP as the desired count of workerplane which is desired
func DelWorkerNodes(client *resources.KsctlClient) ([]string, error) {

	hostnames := client.Cloud.GetHostNameAllWorkerNode()

	if hostnames == nil {
		return nil, log.NewError("hostname is empty")
	}

	currLen := len(hostnames)
	desiredLen := client.Metadata.NoWP
	hostnames = hostnames[desiredLen:]

	if desiredLen < 0 || desiredLen > currLen {
		return nil, log.NewError("not a valid count of wp for down scaling")
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
			return nil, log.NewError(err.Error())
		}
	}

	_, err := client.Cloud.NoOfWorkerPlane(client.Storage, desiredLen, true)
	if err != nil {
		return nil, log.NewError(err.Error())
	}

	return hostnames, nil

}

func CreateHACluster(client *resources.KsctlClient) error {
	var err error
	err = client.Cloud.Name(client.Metadata.ClusterName + "-net").NewNetwork(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	if _, err := client.Cloud.NoOfControlPlane(client.Metadata.NoCP, true); err != nil {
		return log.NewError(err.Error())
	}

	if _, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, true); err != nil {
		return log.NewError(err.Error())
	}

	if _, err := client.Cloud.NoOfDataStore(client.Metadata.NoDS, true); err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-ssh").CreateUploadSSHKeyPair(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-lb").
		Role(consts.RoleLb).
		NewFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-db").
		Role(consts.RoleDs).
		NewFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-cp").
		Role(consts.RoleCp).
		NewFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	err = client.Cloud.Name(client.Metadata.ClusterName + "-fw-wp").
		Role(consts.RoleWp).
		NewFirewall(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
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
			return log.NewError(err.Error())
		}
	}
	for err := range errChanDS {
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	for err := range errChanCP {
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	for err := range errChanWP {
		if err != nil {
			return log.NewError(err.Error())
		}
	}

	return nil
}

func CreateManagedCluster(client *resources.KsctlClient) (bool, bool, error) {

	if client.Metadata.Provider != consts.CloudLocal {
		if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(client.Storage); err != nil {
			return false, false, log.NewError(err.Error())
		}
	}

	managedClient := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed")

	if client.Metadata.Provider != consts.CloudLocal {
		managedClient = managedClient.VMType(client.Metadata.ManagedNodeType)
	}

	externalApps := managedClient.Application(client.Metadata.Applications)

	externalCNI := managedClient.CNI(client.Metadata.CNIPlugin)

	managedClient = managedClient.Version(client.Metadata.K8sVersion)

	if managedClient == nil {
		return externalApps, externalCNI, log.NewError("invalid k8s version")
	}

	if err := managedClient.NewManagedCluster(client.Storage, client.Metadata.NoMP); err != nil {
		return externalApps, externalCNI, log.NewError(err.Error())
	}
	return externalApps, externalCNI, nil
}

func DeleteManagedCluster(client *resources.KsctlClient) error {

	if err := client.Cloud.DelManagedCluster(client.Storage); err != nil {
		return log.NewError(err.Error())
	}

	if client.Metadata.Provider != consts.CloudLocal {
		if err := client.Cloud.DelNetwork(client.Storage); err != nil {
			return log.NewError(err.Error())
		}
	}
	return nil
}
