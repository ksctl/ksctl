package kubernetes

import (
	"os"
	"sync"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/internal/kubernetes"

	"github.com/ksctl/ksctl/internal/k8sdistros"
	k3sPkg "github.com/ksctl/ksctl/internal/k8sdistros/k3s"
	kubeadmPkg "github.com/ksctl/ksctl/internal/k8sdistros/kubeadm"
	"github.com/ksctl/ksctl/internal/storage/external"
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

var log resources.LoggerFactory

func Setup(client *resources.KsctlClient, state *types.StorageDocument) error {
	log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
	log.SetPackageName("kubernetes-controller")

	client.PreBootstrap = k8sdistros.NewPreBootStrap(client.Metadata, state) // NOTE: it needs the

	switch client.Metadata.K8sDistro {
	case consts.K8sK3s:
		client.Bootstrap = k3sPkg.NewClient(client.Metadata, state)
	case consts.K8sKubeadm:
		client.Bootstrap = kubeadmPkg.NewClient(client.Metadata, state)
	default:
		return log.NewError("Invalid k8s provider")
	}
	return nil
}

func ConfigureCluster(client *resources.KsctlClient) (bool, error) {
	waitForPre := &sync.WaitGroup{}

	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, client.Metadata.NoDS)

	waitForPre.Add(1 + client.Metadata.NoDS)

	go func() {
		defer waitForPre.Done()

		err := client.PreBootstrap.ConfigureLoadbalancer(client.Storage)
		if err != nil {
			errChanLB <- err
		}
	}()

	for no := 0; no < client.Metadata.NoDS; no++ {
		go func(i int) {
			defer waitForPre.Done()

			err := client.PreBootstrap.ConfigureDataStore(i, client.Storage)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}
	waitForPre.Wait()
	close(errChanLB)
	close(errChanDS)

	for err := range errChanLB {
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}

	for err := range errChanDS {
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}

	if err := client.Bootstrap.Setup(client.Storage, consts.OperationCreate); err != nil {
		return false, log.NewError(err.Error())
	}

	externalCNI := client.Bootstrap.CNI(client.Metadata.CNIPlugin)

	client.Bootstrap = client.Bootstrap.Version(client.Metadata.K8sVersion)
	if client.Bootstrap == nil {
		return false, log.NewError("invalid version of self-managed k8s cluster")
	}

	// wp[0,N] depends on cp[0]
	err := client.Bootstrap.ConfigureControlPlane(0, client.Storage)
	if err != nil {
		return false, log.NewError(err.Error())
	}

	errChanCP := make(chan error, client.Metadata.NoCP-1)
	errChanWP := make(chan error, client.Metadata.NoWP)

	wg := &sync.WaitGroup{}

	wg.Add(client.Metadata.NoCP - 1 + client.Metadata.NoWP)

	for no := 1; no < client.Metadata.NoCP; no++ {
		go func(i int) {
			defer wg.Done()
			err := client.Bootstrap.ConfigureControlPlane(i, client.Storage)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}
	for no := 0; no < client.Metadata.NoWP; no++ {
		go func(i int) {
			defer wg.Done()
			err := client.Bootstrap.JoinWorkerplane(i, client.Storage)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()

	close(errChanCP)
	close(errChanWP)

	for err := range errChanCP {
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}

	for err := range errChanWP {
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}
	return externalCNI, nil
}

func JoinMoreWorkerPlanes(client *resources.KsctlClient, start, end int) error {

	if err := client.Bootstrap.Setup(client.Storage, consts.OperationGet); err != nil {
		return log.NewError(err.Error())
	}
	client.Bootstrap = client.Bootstrap.Version(client.Metadata.K8sVersion)
	if client.Bootstrap == nil {
		return log.NewError("invalid version of self-managed k8s cluster")
	}
	wg := &sync.WaitGroup{}
	errChan := make(chan error, end-start)

	for no := start; no < end; no++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := client.Bootstrap.JoinWorkerplane(i, client.Storage)
			if err != nil {
				errChan <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return log.NewError(err.Error())
		}
	}

	return nil
}

func DelWorkerPlanes(client *resources.KsctlClient, kubeconfig string, hostnames []string) error {

	kubernetesClient := kubernetes.Kubernetes{
		Metadata:      client.Metadata,
		StorageDriver: client.Storage,
	}
	if err := kubernetesClient.NewKubeconfigClient(kubeconfig); err != nil {
		return log.NewError(err.Error())
	}

	for _, hostname := range hostnames {
		if err := kubernetesClient.DeleteWorkerNodes(hostname); err != nil {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func ApplicationsInCluster(client *resources.KsctlClient, state *types.StorageDocument, op consts.KsctlOperation) error {

	if log == nil {
		log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
		log.SetPackageName("ksctl-distro")
	}

	kubernetesClient := kubernetes.Kubernetes{
		Metadata:      client.Metadata,
		StorageDriver: client.Storage,
	}

	if err := kubernetesClient.NewInClusterClient(); err != nil {
		return log.NewError(err.Error())
	}

	_apps, err := helpers.ToApplicationTempl(client.Metadata.Applications)
	if err != nil {
		return err
	}

	if len(client.Metadata.CNIPlugin) != 0 {
		_cni, err := helpers.ToApplicationTempl([]string{client.Metadata.CNIPlugin})
		if err != nil {
			return err
		}

		if err := kubernetesClient.InstallCNI(_cni[0], state, op); err != nil {
			return err
		}
	}
	if len(client.Metadata.Applications) != 0 {
		return kubernetesClient.Applications(_apps, state, op)
	}
	return nil
}

func InstallAdditionalTools(externalCNI, externalApp bool, client *resources.KsctlClient, state *types.StorageDocument) error {

	if os.Getenv(string(consts.KsctlFakeFlag)) == "1" {
		return nil
	}

	if log == nil {
		log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
		log.SetPackageName("ksctl-distro")
	}

	kubernetesClient := kubernetes.Kubernetes{
		Metadata:      client.Metadata,
		StorageDriver: client.Storage,
	}

	if err := kubernetesClient.NewKubeconfigClient(state.ClusterKubeConfig); err != nil {
		return log.NewError(err.Error())
	}

	if externalCNI {
		var cni string
		if len(client.Metadata.CNIPlugin) == 0 {
			cni = "flannel"
		} else {
			cni = client.Metadata.CNIPlugin
		}

		_cni, err := helpers.ToApplicationTempl([]string{cni})
		if err != nil {
			return err
		}

		if err := kubernetesClient.InstallCNI(_cni[0], state, consts.OperationCreate); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("Done with installing k8s cni")
	}

	if err := installKsctlSpecificApps(client, kubernetesClient, state); err != nil {
		return log.NewError(err.Error())
	}

	if len(client.Metadata.Applications) != 0 && externalApp {
		_apps, err := helpers.ToApplicationTempl(client.Metadata.Applications)
		if err != nil {
			return err
		}
		if err := kubernetesClient.Applications(_apps, state, consts.OperationCreate); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("Done with installing k8s apps")
	}

	log.Success("Done with installing additional k8s tools")
	return nil
}

func installKsctlSpecificApps(client *resources.KsctlClient, kubernetesClient kubernetes.Kubernetes, state *types.StorageDocument) error {

	var (
		exportedData         *resources.StorageStateExportImport
		externalCredEndpoint map[string][]byte
		isExternalStore      bool
	)

	switch client.Metadata.StateLocation {
	case consts.StoreLocal:
		var _err error
		exportedData, _err = client.Storage.Export(map[consts.KsctlSearchFilter]string{
			consts.Cloud:  string(client.Metadata.Provider),
			consts.Name:   client.Metadata.ClusterName,
			consts.Region: client.Metadata.Region,
			consts.ClusterType: func() string {
				if !client.Metadata.IsHA {
					return string(consts.ClusterTypeMang)
				}
				return string(consts.ClusterTypeHa)
			}(),
		})
		if _err != nil {
			return _err
		}

	case consts.StoreExtMongo:
		isExternalStore = true
		var _err error
		externalCredEndpoint, _err = external.HandleCreds(consts.StoreExtMongo)
		if _err != nil {
			return _err
		}
	case consts.StoreK8s:
		// WARN: for now we are not going to transfer state if the ksctl core is already running in one cluster
		// to a new cluster aka (k8s -> k8s)
	}

	if err := kubernetesClient.DeployAgent(client, externalCredEndpoint, isExternalStore); err != nil {
		return log.NewError(err.Error())
	}

	if err := kubernetesClient.DeployRequiredControllers(exportedData, state, isExternalStore); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Done with installing ksctl k8s specific tools")

	return nil
}
