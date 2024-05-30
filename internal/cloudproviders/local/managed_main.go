package local

import (
	"os"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

// DelManagedCluster implements types.CloudFactory.
func (cloud *LocalProvider) DelManagedCluster(storage types.StorageFactory) error {

	cloud.client.NewProvider(log, storage, nil)
	if len(cloud.metadata.tempDirKubeconfig) == 0 {
		var err error
		cloud.metadata.tempDirKubeconfig, err = os.MkdirTemp("", cloud.clusterName+"*")
		if err != nil {
			return err
		}
		if err := os.WriteFile(cloud.metadata.tempDirKubeconfig+helpers.PathSeparator+"kubeconfig",
			[]byte(mainStateDocument.ClusterKubeConfig), 0755); err != nil {
			return err
		}
		defer func() {
			_ = os.RemoveAll(cloud.metadata.tempDirKubeconfig)
		}()
	}

	if err := cloud.client.Delete(cloud.clusterName,
		cloud.metadata.tempDirKubeconfig+helpers.PathSeparator+"kubeconfig"); err != nil {
		return log.NewError(localCtx, "failed to delete cluster", "Reason", err)
	}

	if err := storage.DeleteCluster(); err != nil {
		return err
	}

	return nil
}

func (cloud *LocalProvider) NewManagedCluster(storage types.StorageFactory, noOfNodes int) error {

	vmType := cloud.vmType

	cloud.client.NewProvider(log, storage, nil)

	cni := false
	if consts.KsctlValidCNIPlugin(cloud.metadata.cni) == consts.CNINone {
		cni = true
	}

	withConfig, err := configOption(noOfNodes, cni)
	if err != nil {
		return err
	}

	mainStateDocument.CloudInfra.Local.B.KubernetesVer = cloud.metadata.version
	mainStateDocument.CloudInfra.Local.Nodes = noOfNodes

	mainStateDocument.BootstrapProvider = "kind"
	mainStateDocument.CloudInfra.Local.ManagedNodeSize = vmType

	Wait := 50 * time.Second

	cloud.tempDirKubeconfig, err = os.MkdirTemp("", cloud.clusterName+"*")
	if err != nil {
		return err
	}

	ConfigHandler := func() string {
		path, err := createNecessaryConfigs(cloud.tempDirKubeconfig)
		if err != nil {
			log.Error(localCtx, "rollback Cannot continue 😢")
			err = cloud.DelManagedCluster(storage)
			if err != nil {
				log.Error(localCtx, "failed to perform cleanup", "Reason", err)
				return "" // asumming it never comes here
			}
		}
		return path
	}
	Image := "kindest/node:v" + mainStateDocument.CloudInfra.Local.B.KubernetesVer

	if err := cloud.client.Create(cloud.clusterName, withConfig, Image, Wait, ConfigHandler); err != nil {
		return log.NewError(localCtx, "failed to create cluster", "err", err)
	}

	path := cloud.tempDirKubeconfig + helpers.PathSeparator + "kubeconfig"
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	log.Debug(localCtx, "kubeconfig", "kubeconfigTempPath", path)

	mainStateDocument.ClusterKubeConfig = string(data)
	mainStateDocument.ClusterKubeConfigContext = "kind-" + cloud.clusterName
	mainStateDocument.CloudInfra.Local.B.IsCompleted = true

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	_ = os.RemoveAll(cloud.tempDirKubeconfig) // remove the temp directory

	return nil
}
