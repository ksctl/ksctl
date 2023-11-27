package kubernetes

import (
	k3sPkg "github.com/kubesimplify/ksctl/internal/k8sdistros/k3s"
	kubeadmPkg "github.com/kubesimplify/ksctl/internal/k8sdistros/kubeadm"
	"github.com/kubesimplify/ksctl/internal/k8sdistros/universal"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
)

var log resources.LoggerFactory

func HydrateK8sDistro(client *resources.KsctlClient) error {
	log = logger.NewDefaultLogger(client.Metadata.LogVerbosity, client.Metadata.LogWritter)
	log.SetPackageName("ksctl-distro")

	switch client.Metadata.K8sDistro {
	case consts.K8sK3s:
		client.Distro = k3sPkg.ReturnK3sStruct(client.Metadata)
	case consts.K8sKubeadm:
		client.Distro = kubeadmPkg.ReturnKubeadmStruct(client.Metadata)
	default:
		return log.NewError("Invalid k8s provider")
	}
	return nil
}

func ConfigureCluster(client *resources.KsctlClient) (bool, error) {
	err := client.Distro.ConfigureLoadbalancer(client.Storage)
	if err != nil {
		return false, log.NewError(err.Error())
	}

	for no := 0; no < client.Metadata.NoDS; no++ {
		err := client.Distro.ConfigureDataStore(no, client.Storage)
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}

	externalCNI := client.Distro.CNI(client.Metadata.CNIPlugin)

	client.Distro = client.Distro.Version(client.Metadata.K8sVersion)
	if client.Distro == nil {
		return false, log.NewError("invalid version of self-managed k8s cluster")
	}

	for no := 0; no < client.Metadata.NoCP; no++ {
		err := client.Distro.ConfigureControlPlane(no, client.Storage)
		if err != nil {
			return false, log.NewError(err.Error())
		}
	}

	for no := 0; no < client.Metadata.NoWP; no++ {
		err := client.Distro.JoinWorkerplane(no, client.Storage)
		if err != nil {
			return externalCNI, log.NewError(err.Error())
		}
	}
	return externalCNI, nil
}

func JoinMoreWorkerPlanes(client *resources.KsctlClient, start, end int) error {
	client.Distro = client.Distro.Version(client.Metadata.K8sVersion)
	if client.Distro == nil {
		return log.NewError("invalid version of self-managed k8s cluster")
	}

	for no := start; no < end; no++ {
		err := client.Distro.JoinWorkerplane(no, client.Storage)
		if err != nil {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func DelWorkerPlanes(client *resources.KsctlClient, hostnames []string) error {

	kubeconfigPath, _, err := client.Distro.GetKubeConfig(client.Storage)
	if err != nil {
		return log.NewError(err.Error())
	}

	kubernetesClient := universal.Kubernetes{
		Metadata:      client.Metadata,
		StorageDriver: client.Storage,
	}
	if err := kubernetesClient.ClientInit(kubeconfigPath); err != nil {
		return log.NewError(err.Error())
	}

	for _, hostname := range hostnames {
		if err := kubernetesClient.DeleteWorkerNodes(hostname); err != nil {
			return log.NewError(err.Error())
		}
	}
	return nil
}
