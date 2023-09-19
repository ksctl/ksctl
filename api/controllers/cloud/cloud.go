package cloud

import (
	"fmt"
	"sync"
	"time"

	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	local_pkg "github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
	. "github.com/kubesimplify/ksctl/api/utils/consts"
)

// make it return error
func HydrateCloud(client *resources.KsctlClient, operation KsctlOperation, fakeClient bool) error {
	var err error
	switch client.Metadata.Provider {
	case CLOUD_CIVO:
		if !fakeClient {
			client.Cloud, err = civo_pkg.ReturnCivoStruct(client.Metadata, civo_pkg.ProvideClient)
		} else {
			client.Cloud, err = civo_pkg.ReturnCivoStruct(client.Metadata, civo_pkg.ProvideMockCivoClient)
		}

		if err != nil {
			return fmt.Errorf("[cloud] " + err.Error())
		}
	case CLOUD_AZURE:
		if !fakeClient {
			client.Cloud, err = azure_pkg.ReturnAzureStruct(client.Metadata, azure_pkg.ProvideClient)
		} else {
			client.Cloud, err = azure_pkg.ReturnAzureStruct(client.Metadata, azure_pkg.ProvideMockClient)
		}

		if err != nil {
			return fmt.Errorf("[cloud] " + err.Error())
		}
	case CLOUD_LOCAL:
		client.Cloud, err = local_pkg.ReturnLocalStruct(client.Metadata)
		if err != nil {
			return fmt.Errorf("[cloud] " + err.Error())
		}
	default:
		return fmt.Errorf("Invalid Cloud provider")
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

func DeleteHACluster(client *resources.KsctlClient) error {

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

			err := client.Cloud.Role(ROLE_WP).DelVM(client.Storage, no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	for no := 0; no < noCP; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(ROLE_CP).DelVM(client.Storage, no)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}
	for no := 0; no < noDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(ROLE_DS).DelVM(client.Storage, no)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}

	go func() {
		defer wg.Done()

		err := client.Cloud.Role(ROLE_LB).DelVM(client.Storage, 0)
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

	err = client.Cloud.Role(ROLE_DS).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(ROLE_CP).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(ROLE_WP).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(ROLE_LB).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

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
func AddWorkerNodes(client *resources.KsctlClient) (int, error) {

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

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", client.ClusterName, no)).
				Role(ROLE_WP).
				VMType(client.WorkerPlaneNodeType).
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
func DelWorkerNodes(client *resources.KsctlClient) ([]string, error) {

	hostnames := client.Cloud.GetHostNameAllWorkerNode()

	if hostnames == nil {
		return nil, fmt.Errorf("[cloud] hostname is empty")
	}

	currLen := len(hostnames)
	desiredLen := client.Metadata.NoWP
	hostnames = hostnames[desiredLen:]

	if desiredLen < 0 || desiredLen > currLen {
		return nil, fmt.Errorf("[cloud] not a valid count of wp for down scaling")
	}

	wg := &sync.WaitGroup{}
	errChanWP := make(chan error, currLen-desiredLen)

	for no := desiredLen; no < currLen; no++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Role(ROLE_WP).DelVM(client.Storage, no)
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

func CreateHACluster(client *resources.KsctlClient) error {
	var err error
	err = client.Cloud.Name(client.ClusterName + "-net").NewNetwork(client.Storage)
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

	err = client.Cloud.Name(client.ClusterName + "-ssh").CreateUploadSSHKeyPair(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-lb").
		Role(ROLE_LB).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-db").
		Role(ROLE_DS).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-cp").
		Role(ROLE_CP).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-wp").
		Role(ROLE_WP).
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

		err := client.Cloud.Name(client.ClusterName+"-vm-lb").
			Role(ROLE_LB).
			VMType(client.LoadBalancerNodeType).
			Visibility(true).
			NewVM(client.Storage, 0)
		if err != nil {
			errChanLB <- err
		}
	}()

	for no := 0; no < client.Metadata.NoDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-db-%d", client.ClusterName, no)).
				Role(ROLE_DS).
				VMType(client.DataStoreNodeType).
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

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-cp-%d", client.ClusterName, no)).
				Role(ROLE_CP).
				VMType(client.ControlPlaneNodeType).
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

			err := client.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", client.ClusterName, no)).
				Role(ROLE_WP).
				VMType(client.WorkerPlaneNodeType).
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

func CreateManagedCluster(client *resources.KsctlClient) error {

	if client.Metadata.Provider != CLOUD_LOCAL {
		if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(client.Storage); err != nil {
			return err
		}
	}

	managedClient := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed")

	if client.Metadata.Provider != CLOUD_LOCAL {
		managedClient = managedClient.VMType(client.Metadata.ManagedNodeType)
	}

	if client.Cloud.SupportForApplications() {
		managedClient = managedClient.Application(client.Metadata.Applications)
	}

	if client.Cloud.SupportForCNI() {
		managedClient = managedClient.CNI(client.Metadata.CNIPlugin)
	}

	managedClient = managedClient.Version(client.Metadata.K8sVersion)

	if managedClient == nil {
		return fmt.Errorf("[azure] invalid version")
	}

	if err := managedClient.NewManagedCluster(client.Storage, client.Metadata.NoMP); err != nil {
		return err
	}
	return nil
}

func DeleteManagedCluster(client *resources.KsctlClient) error {

	if err := client.Cloud.DelManagedCluster(client.Storage); err != nil {
		return err
	}

	if client.Metadata.Provider != CLOUD_LOCAL {
		if err := client.Cloud.DelNetwork(client.Storage); err != nil {
			return err
		}
	}
	return nil
}
