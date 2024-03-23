package kubernetes

import (
	"os"
	"strings"

	"github.com/ksctl/ksctl/internal/k8sdistros"
	k3sPkg "github.com/ksctl/ksctl/internal/k8sdistros/k3s"
	kubeadmPkg "github.com/ksctl/ksctl/internal/k8sdistros/kubeadm"
	"github.com/ksctl/ksctl/internal/k8sdistros/universal"
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
	//waitForPre := &sync.WaitGroup{}

	if err := client.PreBootstrap.ConfigureLoadbalancer(client.Storage); err != nil {
		return false, log.NewError(err.Error())
	}
	//go func() {
	//	waitForPre.Add(1)
	//	defer waitForPre.Done()
	//	_ = client.PreBootstrap.ConfigureLoadbalancer(client.Storage)
	//	// if err != nil {
	//	// 	return false, log.NewError(err.Error())
	//	// }
	//}()

	for no := 0; no < client.Metadata.NoDS; no++ {
		if err := client.PreBootstrap.ConfigureDataStore(no, client.Storage); err != nil {
			return false, log.NewError(err.Error())
		}
		//waitForPre.Add(1)
		//go func(i int) {
		//	defer waitForPre.Done()
		//	_ = client.PreBootstrap.ConfigureDataStore(i, client.Storage)
		//}(no)
	}
	//waitForPre.Wait()

	// TODO we can try to make it FIXME aka merge it
	// WARN please fix it
	if err := client.Bootstrap.Setup(client.Storage, consts.OperationStateCreate); err != nil {
		return false, log.NewError(err.Error())
	}

	externalCNI := client.Bootstrap.CNI(client.Metadata.CNIPlugin)

	client.Bootstrap = client.Bootstrap.Version(client.Metadata.K8sVersion)
	if client.Bootstrap == nil {
		return false, log.NewError("invalid version of self-managed k8s cluster")
	}

	//wg := &sync.WaitGroup{}

	// wp[0,N] depends on cp[0]
	err := client.Bootstrap.ConfigureControlPlane(0, client.Storage)
	if err != nil {
		return false, log.NewError(err.Error())
	}

	for no := 1; no < client.Metadata.NoCP; no++ {
		err := client.Bootstrap.ConfigureControlPlane(no, client.Storage)
		if err != nil {
			return false, log.NewError(err.Error())
		}
		//wg.Add(1)
		//go func(i int) {
		//	defer wg.Done()
		//	_ = client.Bootstrap.ConfigureControlPlane(i, client.Storage)
		//}(no)
	}
	for no := 0; no < client.Metadata.NoWP; no++ {
		err := client.Bootstrap.JoinWorkerplane(no, client.Storage)
		if err != nil {
			return externalCNI, log.NewError(err.Error())
		}
		//wg.Add(1)
		//go func(i int) {
		//	defer wg.Done()
		//	_ = client.Bootstrap.JoinWorkerplane(i, client.Storage)
		//}(no)
	}
	//wg.Wait()

	return externalCNI, nil
}

func JoinMoreWorkerPlanes(client *resources.KsctlClient, start, end int) error {

	// TODO we can try to make it FIXME aka merge it
	// WARN please fix it
	if err := client.Bootstrap.Setup(client.Storage, consts.OperationStateGet); err != nil {
		return log.NewError(err.Error())
	}
	client.Bootstrap = client.Bootstrap.Version(client.Metadata.K8sVersion)
	if client.Bootstrap == nil {
		return log.NewError("invalid version of self-managed k8s cluster")
	}
	//wg := &sync.WaitGroup{}

	for no := start; no < end; no++ {
		//wg.Add(1)
		//go func(i int) {
		//	defer wg.Done()
		//	_ = client.Bootstrap.JoinWorkerplane(i, client.Storage)
		//}(no)

		if err := client.Bootstrap.JoinWorkerplane(no, client.Storage); err != nil {
			return log.NewError(err.Error())
		}
	}
	//wg.Wait()
	return nil
}

func DelWorkerPlanes(client *resources.KsctlClient, kubeconfig string, hostnames []string) error {

	kubernetesClient := universal.Kubernetes{
		Metadata:      client.Metadata,
		StorageDriver: client.Storage,
	}
	if err := kubernetesClient.ClientInit(kubeconfig); err != nil {
		return log.NewError(err.Error())
	}

	for _, hostname := range hostnames {
		if err := kubernetesClient.DeleteWorkerNodes(hostname); err != nil {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func InstallAdditionalTools(kubeconfig string, externalCNI, externalApp bool, client *resources.KsctlClient, state *types.StorageDocument) error {

	if log == nil {
		log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
		log.SetPackageName("ksctl-distro")
	}

	var kubernetesClient universal.Kubernetes

	if externalCNI || (len(client.Metadata.Applications) != 0 && externalApp) {
		kubernetesClient = universal.Kubernetes{
			Metadata:      client.Metadata,
			StorageDriver: client.Storage,
		}
		if err := kubernetesClient.ClientInit(kubeconfig); err != nil {
			return log.NewError(err.Error())
		}
	}

	if externalCNI {
		if err := kubernetesClient.InstallCNI(client.Metadata.CNIPlugin); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("Done with installing k8s cni")
	}

	if len(client.Metadata.Applications) != 0 && externalApp {
		apps := strings.Split(client.Metadata.Applications, ",")
		if err := kubernetesClient.InstallApplications(apps); err != nil {
			return log.NewError(err.Error())
		}

		log.Success("Done with installing k8s apps")
	}

	if err := installKsctlSpecificApps(client, kubeconfig, kubernetesClient, state); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Done with installing additional k8s tools")
	return nil
}

func installKsctlSpecificApps(client *resources.KsctlClient, kubeconfig string, kubernetesClient universal.Kubernetes, state *types.StorageDocument) error {

	var cloudSecret map[string][]byte
	var err error
	cloudSecret, err = client.Cloud.GetSecretTokens(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	////// EXPERIMENTAL Features //////
	if len(os.Getenv(string(consts.KsctlFeatureFlagHaAutoscale))) > 0 {

		if err = kubernetesClient.KsctlConfigForController(kubeconfig, state, cloudSecret); err != nil {
			return log.NewError(err.Error())
		}
	}

	log.Success("Done with installing ksctl k8s specific tools")

	return nil
}
