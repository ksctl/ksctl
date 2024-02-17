package k3s

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
	"github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

var (
	mainStateDocument *types.StorageDocument
	log               resources.LoggerFactory
)

type K3sDistro struct {
	K3sVer string
	Cni    string
	// it will be used for SSH
	SSHInfo helpers.SSHCollection
}

func ReturnK3sStruct(meta resources.Metadata, state *types.StorageDocument) *K3sDistro {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName("k3s")

	mainStateDocument = state

	return &K3sDistro{
		SSHInfo: &helpers.SSHPayload{},
	}
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
sudo cat /etc/rancher/k3s/k3s.yaml`
}

// InitState implements resources.DistroFactory.
// try to achieve deepCopy
func (k3s *K3sDistro) InitState(cloudState cloud.CloudResourceState, storage resources.StorageFactory, operation consts.KsctlOperation) error {

	if operation == consts.OperationStateCreate {
		mainStateDocument.K8sBootstrap = &types.KubernetesBootstrapState{K3s: &types.StateConfigurationK3s{}}
	}

	mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.ControlPlanes = cloudState.IPv4ControlPlanes
	mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.ControlPlanes = cloudState.PrivateIPv4ControlPlanes

	mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.DataStores = cloudState.IPv4DataStores
	mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.DataStores = cloudState.PrivateIPv4DataStores

	mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.WorkerPlanes = cloudState.IPv4WorkerPlanes

	mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.LoadBalancer = cloudState.IPv4LoadBalancer
	mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.LoadBalancer = cloudState.PrivateIPv4LoadBalancer
	mainStateDocument.K8sBootstrap.K3s.B.SSHInfo = cloudState.SSHState

	k3s.SSHInfo.PrivateKey(mainStateDocument.K8sBootstrap.K3s.B.SSHInfo.PrivateKey)
	k3s.SSHInfo.Username(mainStateDocument.K8sBootstrap.K3s.B.SSHInfo.UserName)

	err := storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError("failed to Initialized state from Cloud reason: %v", err)
	}
	log.Debug("Printing", "k3sState", mainStateDocument)

	log.Print("Initialized state from Cloud")
	return nil
}

func (k3s *K3sDistro) Version(ver string) resources.DistroFactory {
	if isValidK3sVersion(ver) {
		// valid
		k3s.K3sVer = fmt.Sprintf("v%s+k3s1", ver)
		log.Debug("Printing", "k3s.K3sVer", k3s.K3sVer)
		return k3s
	}
	return nil
}

func (k3s *K3sDistro) CNI(cni string) (externalCNI bool) {
	log.Debug("Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case consts.CNIFlannel, "":
		k3s.Cni = string(consts.CNIFlannel)
	default:
		k3s.Cni = string(consts.CNINone)
		return true
	}

	return false
}
