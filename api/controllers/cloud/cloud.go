package cloud

import (
	"fmt"
	"time"

	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	local_pkg "github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// make it return error
func HydrateCloud(client *resources.KsctlClient, operation string) error {
	var err error
	switch client.Metadata.Provider {
	case "civo":
		client.Cloud, err = civo_pkg.ReturnCivoStruct(client.Metadata)
		if err != nil {
			return fmt.Errorf("[cloud] " + err.Error())
		}
	case "azure":
		client.Cloud = azure_pkg.ReturnAzureStruct(client.Metadata)
	case "local":
		client.Cloud = local_pkg.ReturnLocalStruct(client.Metadata)
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

	for i := 0; i < noWP; i++ {
		err = client.Cloud.Role(utils.ROLE_WP).DelVM(client.Storage, i)
		if err != nil {
			return err
		}
	}
	pauseOperation(5)

	for i := 0; i < noCP; i++ {
		err = client.Cloud.Role(utils.ROLE_CP).DelVM(client.Storage, i)
		if err != nil {
			return err
		}
	}
	pauseOperation(5)

	for i := 0; i < noDS; i++ {
		err = client.Cloud.Role(utils.ROLE_DS).DelVM(client.Storage, i)
		if err != nil {
			return err
		}
	}
	pauseOperation(5)

	err = client.Cloud.Role(utils.ROLE_LB).DelVM(client.Storage, 0)
	if err != nil {
		return err
	}

	pauseOperation(5)

	err = client.Cloud.Role(utils.ROLE_DS).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(utils.ROLE_CP).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(utils.ROLE_WP).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Role(utils.ROLE_LB).DelFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.DelSSHKeyPair(client.Storage)
	if err != nil {
		return err
	}

	// last one to delete is network
	err = client.Cloud.DelNetwork(client.Storage)
	if err != nil {
		return err
	}

	return nil
}

// the user provides the desired no of workerplane not the no of workerplanes to be added
func AddWorkerNodes(client *resources.KsctlClient) (int, error) {

	currWP, err := client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, false)
	if err != nil {
		return -1, err
	}

	_, err = client.Cloud.NoOfWorkerPlane(client.Storage, client.Metadata.NoWP, true)
	if err != nil {
		return -1, err
	}

	for no := currWP; no < client.Metadata.NoWP; no++ {
		name := client.ClusterName + "-vm-wp"
		_ = client.Cloud.Name(fmt.Sprintf("%s-%d", name, no)).
			Role(utils.ROLE_WP).
			VMType(client.WorkerPlaneNodeType).
			Visibility(true).
			NewVM(client.Storage, no)
	}

	// workerplane created
	return currWP, nil
}

// it uses the noWP as the desired count of workerplane which is desired
func DelWorkerNodes(client *resources.KsctlClient) ([]string, error) {

	hostnames := client.Cloud.GetHostNameAllWorkerNode()

	if hostnames == nil {
		return nil, fmt.Errorf("[cloud] hostname is empty")
	}

	currLen := len(hostnames)
	desiredLen := client.Metadata.NoWP
	hostnames = hostnames[:desiredLen]

	if desiredLen < 0 || desiredLen > currLen {
		return nil, fmt.Errorf("[cloud] not a valid count of wp for down scaling")
	}

	for i := desiredLen; i < currLen; i++ {
		err := client.Cloud.Role(utils.ROLE_DS).DelVM(client.Storage, i)
		if err != nil {
			return nil, err
		}
	}
	pauseOperation(5)

	_, err := client.Cloud.NoOfWorkerPlane(client.Storage, desiredLen, true)
	if err != nil {
		return nil, err
	}

	return hostnames, nil

}

func CreateHACluster(client *resources.KsctlClient) error {
	err := client.Cloud.Name(client.ClusterName + "-net").NewNetwork(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-ssh").CreateUploadSSHKeyPair(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-lb").
		Role(utils.ROLE_LB).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-db").
		Role(utils.ROLE_DS).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-cp").
		Role(utils.ROLE_CP).
		NewFirewall(client.Storage)
	if err != nil {
		return err
	}

	err = client.Cloud.Name(client.ClusterName + "-fw-wp").
		Role(utils.ROLE_WP).
		NewFirewall(client.Storage)
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

	_ = client.Cloud.Name(client.ClusterName+"-vm-lb").
		Role(utils.ROLE_LB).
		VMType(client.LoadBalancerNodeType).
		Visibility(true).
		NewVM(client.Storage, 0)

	for no := 0; no < client.Metadata.NoDS; no++ {
		name := client.ClusterName + "-vm-db"
		_ = client.Cloud.Name(fmt.Sprintf("%s-%d", name, no)).
			Role(utils.ROLE_DS).
			VMType(client.DataStoreNodeType).
			Visibility(true).
			NewVM(client.Storage, no)
	}

	for no := 0; no < client.Metadata.NoCP; no++ {
		name := client.ClusterName + "-vm-cp"
		_ = client.Cloud.Name(fmt.Sprintf("%s-%d", name, no)).
			Role(utils.ROLE_CP).
			VMType(client.ControlPlaneNodeType).
			Visibility(true).
			NewVM(client.Storage, no)
	}

	for no := 0; no < client.Metadata.NoWP; no++ {
		name := client.ClusterName + "-vm-wp"
		_ = client.Cloud.Name(fmt.Sprintf("%s-%d", name, no)).
			Role(utils.ROLE_WP).
			VMType(client.WorkerPlaneNodeType).
			Visibility(true).
			NewVM(client.Storage, no)
	}
	return nil
}

func CreateManagedCluster(client *resources.KsctlClient) error {
	if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(client.Storage); err != nil {
		// need to verify wrt to other providers wrt network creation
		return err
	}

	managedClient := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed").
		VMType(client.Metadata.ManagedNodeType)

	if client.Cloud.SupportForApplications() {
		managedClient = managedClient.Application(client.Metadata.Applications)
	}

	if client.Cloud.SupportForCNI() {
		managedClient = managedClient.CNI(client.Metadata.CNIPlugin)
	}

	managedClient = managedClient.Version(client.Metadata.K8sVersion) // for kubernetes version

	if managedClient == nil {
		client.Storage.Logger().Err("Invalid version")
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

	if err := client.Cloud.DelNetwork(client.Storage); err != nil {
		return err
	}
	return nil
}
