package local

import (
	"time"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
	"sigs.k8s.io/kind/pkg/cluster"
)

// DelManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) DelManagedCluster(storage resources.StorageFactory) error {

	logger := CustomLogger{log}
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	if err := provider.Delete(cloud.ClusterName, utils.GetPath(UtilClusterPath, CloudLocal, ClusterTypeMang, cloud.ClusterName, KUBECONFIG)); err != nil {
		return log.NewError("failed to delete cluster %v", err)
	}
	printKubeconfig(storage, OperationStateDelete, cloud.ClusterName)

	if err := storage.Path(utils.GetPath(UtilClusterPath, CloudLocal, ClusterTypeMang, cloud.ClusterName)).DeleteDir(); err != nil {
		return log.NewError(err.Error())
	}

	return nil
}

// NewManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	logger := CustomLogger{log}
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	cni := false
	if KsctlValidCNIPlugin(cloud.Metadata.Cni) == CNINone {
		cni = true
	}

	withConfig, err := configOption(noOfNodes, cni)
	if err != nil {
		return log.NewError(err.Error())
	}

	localState.Version = cloud.Metadata.Version
	localState.Nodes = noOfNodes

	Wait := 50 * time.Second
	if err := provider.Create(
		cloud.ClusterName,
		withConfig,
		cluster.CreateWithNodeImage("kindest/node:v"+localState.Version),
		// cluster.CreateWithRetain(flags.Retain),
		cluster.CreateWithWaitForReady(Wait),
		cluster.CreateWithKubeconfigPath(func() string {
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
		}()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return log.NewError("failed to create cluster", "err", err)
	}

	printKubeconfig(storage, OperationStateCreate, cloud.ClusterName)
	return nil
}

func (obj *LocalProvider) GetKubeconfigPath() string {
	return utils.GetPath(UtilClusterPath, CloudLocal, ClusterTypeMang, obj.ClusterName, KUBECONFIG)
}
