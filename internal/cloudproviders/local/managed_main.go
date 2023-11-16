package local

import (
	"time"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// DelManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) DelManagedCluster(storage resources.StorageFactory) error {

	cloud.client.NewProvider(log, storage, nil)

	if err := cloud.client.Delete(cloud.ClusterName, utils.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, cloud.ClusterName, KUBECONFIG)); err != nil {
		return log.NewError("failed to delete cluster %v", err)
	}
	printKubeconfig(storage, consts.OperationStateDelete, cloud.ClusterName)

	if err := storage.Path(utils.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, cloud.ClusterName)).DeleteDir(); err != nil {
		return log.NewError(err.Error())
	}

	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	cloud.client.NewProvider(log, storage, nil)

	cni := false
	if consts.KsctlValidCNIPlugin(cloud.Metadata.Cni) == consts.CNINone {
		cni = true
	}

	withConfig, err := configOption(noOfNodes, cni)
	if err != nil {
		return log.NewError(err.Error())
	}

	localState.Version = cloud.Metadata.Version
	localState.Nodes = noOfNodes

	Wait := 50 * time.Second
	ConfigHandler := func() string {
		path, err := createNecessaryConfigs(storage, cloud.ClusterName)
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
	Image := "kindest/node:v" + localState.Version

	if err := cloud.client.Create(cloud.ClusterName, withConfig, Image, Wait, ConfigHandler); err != nil {
		return log.NewError("failed to create cluster", "err", err)
	}

	printKubeconfig(storage, consts.OperationStateCreate, cloud.ClusterName)
	return nil
}

func (obj *LocalProvider) GetKubeconfigPath() string {
	return utils.GetPath(consts.UtilClusterPath, consts.CloudLocal, consts.ClusterTypeMang, obj.ClusterName, KUBECONFIG)
}
