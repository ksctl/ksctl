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
	if len(cloud.Metadata.tempDirKubeconfig) == 0 {
		var err error
		cloud.Metadata.tempDirKubeconfig, err = os.MkdirTemp("", cloud.ClusterName+"*")
		if err != nil {
			return err
		}
		if err := os.WriteFile(cloud.Metadata.tempDirKubeconfig+helpers.PathSeparator+"kubeconfig",
			[]byte(mainStateDocument.ClusterKubeConfig), 0755); err != nil {
			return err
		}
		defer func() {
			_ = os.RemoveAll(cloud.Metadata.tempDirKubeconfig)
		}()
	}

	if err := cloud.client.Delete(cloud.ClusterName,
		cloud.Metadata.tempDirKubeconfig+helpers.PathSeparator+"kubeconfig"); err != nil {
		return log.NewError("failed to delete cluster %v", err)
	}

	if err := storage.DeleteCluster(); err != nil {
		return log.NewError(err.Error())
	}

	return nil
}

// NewManagedCluster implements types.CloudFactory.
func (cloud *LocalProvider) NewManagedCluster(storage types.StorageFactory, noOfNodes int) error {

	cloud.client.NewProvider(log, storage, nil)

	cni := false
	if consts.KsctlValidCNIPlugin(cloud.Metadata.Cni) == consts.CNINone {
		cni = true
	}

	withConfig, err := configOption(noOfNodes, cni)
	if err != nil {
		return log.NewError(err.Error())
	}

	mainStateDocument.CloudInfra.Local.B.KubernetesVer = cloud.Metadata.Version
	mainStateDocument.CloudInfra.Local.Nodes = noOfNodes

	Wait := 50 * time.Second

	cloud.tempDirKubeconfig, err = os.MkdirTemp("", cloud.ClusterName+"*")
	if err != nil {
		return err
	}

	ConfigHandler := func() string {
		path, err := createNecessaryConfigs(cloud.tempDirKubeconfig)
		if err != nil {
			log.Error("rollback Cannot continue 😢")
			err = cloud.DelManagedCluster(storage)
			if err != nil {
				log.Error(err.Error())
				return "" // asumming it never comes here
			}
		}
		return path
	}
	Image := "kindest/node:v" + mainStateDocument.CloudInfra.Local.B.KubernetesVer

	if err := cloud.client.Create(cloud.ClusterName, withConfig, Image, Wait, ConfigHandler); err != nil {
		return log.NewError("failed to create cluster", "err", err)
	}

	path := cloud.tempDirKubeconfig + helpers.PathSeparator + "kubeconfig"
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	log.Debug("kubeconfig", "kubeconfigTempPath", path)

	mainStateDocument.ClusterKubeConfig = string(data)
	mainStateDocument.CloudInfra.Local.B.IsCompleted = true

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	_ = os.RemoveAll(cloud.tempDirKubeconfig) // remove the temp directory

	return nil
}
