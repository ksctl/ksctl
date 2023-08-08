package kubernetes

import (
	"fmt"

	k3s_pkg "github.com/kubesimplify/ksctl/api/k8s_distro/k3s"
	kubeadm_pkg "github.com/kubesimplify/ksctl/api/k8s_distro/kubeadm"
	"github.com/kubesimplify/ksctl/api/resources"
)

func HydrateK8sDistro(client *resources.KsctlClient) error {
	switch client.Metadata.K8sDistro {
	case "k3s":
		client.Distro = k3s_pkg.ReturnK3sStruct()
	case "kubeadm":
		client.Distro = kubeadm_pkg.ReturnKubeadmStruct()
	default:
		return fmt.Errorf("[kubernetes] Invalid k8s provider")
	}
	return nil
}

func ConfigureCluster(client *resources.KsctlClient) error {
	err := client.Distro.ConfigureLoadbalancer(client.Storage)
	if err != nil {
		return err
	}

	for no := 0; no < client.Metadata.NoDS; no++ {
		err := client.Distro.ConfigureDataStore(no, client.Storage)
		if err != nil {
			return err
		}
	}
	for no := 0; no < client.Metadata.NoCP; no++ {
		err := client.Distro.ConfigureControlPlane(no, client.Storage)
		if err != nil {
			return err
		}
	}

	for no := 0; no < int(client.Metadata.NoWP); no++ {
		err := client.Distro.JoinWorkerplane(no, client.Storage)
		if err != nil {
			return err
		}
	}
	return nil
}

// its [start, end)
func JoinMoreWorkerPlanes(client *resources.KsctlClient, start, end int) error {
	for no := start; no < end; no++ {
		err := client.Distro.JoinWorkerplane(no, client.Storage)
		if err != nil {
			return err
		}
	}
	return nil
}
