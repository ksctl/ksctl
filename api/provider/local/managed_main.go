package local

import (
	"fmt"
	"time"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"
)

// DelManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) DelManagedCluster(storage resources.StorageFactory) error {

	logger := CustomLogger{StorageDriver: storage}
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	if err := provider.Delete(cloud.ClusterName, utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, cloud.ClusterName, KUBECONFIG)); err != nil {
		return fmt.Errorf("[local] failed to delete cluster %v", err)
	}
	printKubeconfig(storage, utils.OPERATION_STATE_DELETE, cloud.ClusterName)

	return storage.Path(utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_LOCAL, cloud.ClusterName)).DeleteDir()
}

// NewManagedCluster implements resources.CloudFactory.
func (cloud *LocalProvider) NewManagedCluster(storage resources.StorageFactory, noOfNodes int) error {

	logger := CustomLogger{StorageDriver: storage}
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	withConfig, err := configOption(noOfNodes)
	if err != nil {
		return err
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
				storage.Logger().Err("[rollback] Cannot continue 😢")
				err = cloud.DelManagedCluster(storage)
				if err != nil {
					panic(err)
				}
			}
			return path
		}()),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return errors.Wrap(err, "[local] failed to create cluster")
	}

	printKubeconfig(storage, utils.OPERATION_STATE_CREATE, cloud.ClusterName)
	return nil
}
