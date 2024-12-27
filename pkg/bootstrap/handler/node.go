package handler

import (
	"strings"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
)

func (k *K8sClusterClient) DeleteWorkerNodes(nodeName string) error {

	// TODO: Need to added step to drain the node before deleting it!

	nodes, err := k.k8sClient.NodesList()
	if err != nil {
		return err
	}

	kNodeName := ""
	for _, node := range nodes.Items {
		k.l.Debug(k.ctx, "string compariazion", "nodeToDelete", nodeName, "kubernetesNodeName", node.Name)
		if strings.HasPrefix(node.Name, nodeName) {
			kNodeName = node.Name
			break
		}
	}

	if len(kNodeName) == 0 {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			k.l.NewError(k.ctx, "node not found!"),
		)
	}
	err = k.k8sClient.NodeDelete(kNodeName)
	if err != nil {
		return err
	}
	k.l.Success(k.ctx, "Deleted Node", "name", kNodeName)
	return nil
}
