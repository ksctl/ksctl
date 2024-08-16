package kubernetes

import (
	"strings"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"k8s.io/client-go/util/retry"
)

func (k *K8sClusterClient) AppPerformPostUninstall(
	apps types.KsctlApp,
	state *storageTypes.StorageDocument,
) error {
	if strings.HasPrefix(apps.StackName, "production-") {
		switch apps.StackName {
		case string(metadata.SpinKubeProductionStackID):
			return k.spinKubePreUninstall()
		}
	}

	return nil
}

func (k *K8sClusterClient) AppPerformPostInstall(
	apps types.KsctlApp,
	state *storageTypes.StorageDocument,
) error {

	if strings.HasPrefix(apps.StackName, "production-") {
		switch apps.StackName {
		case string(metadata.SpinKubeProductionStackID):
			return k.spinKubePostInstall()
		}
	}

	return nil
}

func (k *K8sClusterClient) spinKubePostInstall() error {
	nodes, err := k.k8sClient.NodesList(kubernetesCtx, log)
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		log.Print(kubernetesCtx, node.Name)
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
			}
			node.Annotations["kwasm.sh/kwasm-node"] = "true"

			_, updateErr := k.k8sClient.NodeUpdate(kubernetesCtx, log, &node)
			return updateErr
		})
		if retryErr != nil {
			log.Warn(kubernetesCtx, "Failed to annotate nodes", "targetNodeName", node.Name, "error", retryErr)
		}
		log.Success(kubernetesCtx, "Annotated node", "targetNodeName", node.Name)
	}

	return nil
}

func (k *K8sClusterClient) spinKubePreUninstall() error {
	nodes, err := k.k8sClient.NodesList(kubernetesCtx, log)
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		log.Print(kubernetesCtx, node.Name)
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			if node.Annotations != nil {
				if _, ok := node.Annotations["kwasm.sh/kwasm-node"]; ok {
					delete(node.Annotations, "kwasm.sh/kwasm-node")
				}
			}

			_, updateErr := k.k8sClient.NodeUpdate(kubernetesCtx, log, &node)
			return updateErr
		})
		if retryErr != nil {
			log.Warn(kubernetesCtx, "Failed to remove annotatation from node", "targetNodeName", node.Name, "error", retryErr)
		}
		log.Success(kubernetesCtx, "Removed Annotation from node", "targetNodeName", node.Name)
	}

	return nil
}
