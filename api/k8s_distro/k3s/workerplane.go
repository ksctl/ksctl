package k3s

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3sDistro) JoinWorkerplane(idx int, storage resources.StorageFactory) error {
	path := utils.GetPath(utils.CLUSTER_PATH, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err := saveStateHelper(storage, path)
	if err != nil {
		return err
	}

	err = k3s.SSHInfo.Flag(utils.EXEC_WITHOUT_OUTPUT).Script(
		scriptWP(k8sState.PrivateIPs.Loadbalancer, k8sState.K3sToken)).
		IPv4(k8sState.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return err
	}

	storage.Logger().Success("[k3s] configured WorkerPlane")

	return nil
}

//
//// DestroyWorkerPlane implements resources.DistroFactory.
//func (*K3sDistro) DestroyWorkerPlane(storage resources.StorageFactory) ([]string, error) {
//
//	return nil, nil
//}

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://%s:6443
`, token, privateIPlb)
}
