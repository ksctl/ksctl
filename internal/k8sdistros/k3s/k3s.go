package k3s

import (
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

var (
	mainStateDocument *types.StorageDocument
	log               resources.LoggerFactory
)

type K3s struct {
	K3sVer string
	Cni    string
	mu     *sync.Mutex
}

func NewClient(m resources.Metadata, state *types.StorageDocument) resources.KubernetesBootstrap {
	log = logger.NewDefaultLogger(m.LogVerbosity, m.LogWritter)
	log.SetPackageName("k3s")

	mainStateDocument = state
	return &K3s{mu: &sync.Mutex{}}
}

func (k3s *K3s) Setup(storage resources.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationStateCreate {
		mainStateDocument.K8sBootstrap.K3s = &types.StateConfigurationK3s{}
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
sudo cat /etc/rancher/k3s/k3s.yaml`
}

func (k3s *K3s) Version(ver string) resources.KubernetesBootstrap {
	if isValidK3sVersion(ver) {
		// valid
		k3s.K3sVer = fmt.Sprintf("v%s+k3s1", ver)
		log.Debug("Printing", "k3s.K3sVer", k3s.K3sVer)
		return k3s
	}
	return nil
}

func (k3s *K3s) CNI(cni string) (externalCNI bool) {
	log.Debug("Printing", "cni", cni)
	switch consts.KsctlValidCNIPlugin(cni) {
	case consts.CNIFlannel, "":
		k3s.Cni = string(consts.CNIFlannel)
		return false

	default:
		// this tells us that CNI should be installed via the k8s client
		k3s.Cni = string(consts.CNINone)
		return true
	}
}
