// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
