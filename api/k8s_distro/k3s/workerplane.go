package k3s

import (
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// JoinWorkerplane implements resources.DistroFactory.
func (k3s *K3sDistro) JoinWorkerplane(idx int, storage resources.StorageFactory) error {
	err := k3s.SSHInfo.Flag(utils.EXEC_WITHOUT_OUTPUT).Script(
		scriptWP(k8sState.PrivateIPs.Loadbalancer, k8sState.K3sToken)).
		IPv4(k8sState.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return err
	}
	storage.Logger().Success("[k3s] configured WorkerPlane")

	return nil
}

// DestroyWorkerPlane implements resources.DistroFactory.
func (*K3sDistro) DestroyWorkerPlane(storage resources.StorageFactory) error {
	return nil
}

func scriptWP(privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://%s:6443
`, token, privateIPlb)
}
