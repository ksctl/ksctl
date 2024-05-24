package k3s

import (
	"context"
	"fmt"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
	k3sCtx            context.Context
)

type K3s struct {
	K3sVer string
	Cni    string
	mu     *sync.Mutex
}

func NewClient(
	parentCtx context.Context,
	parentLog types.LoggerFactory,
	state *storageTypes.StorageDocument) *K3s {
	k3sCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.K8sK3s))
	log = parentLog

	mainStateDocument = state
	return &K3s{mu: &sync.Mutex{}}
}

func (k3s *K3s) Setup(storage types.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap.K3s = &storageTypes.StateConfigurationK3s{}
		mainStateDocument.BootstrapProvider = consts.K8sK3s
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return err
	}
	return nil
}

func scriptKUBECONFIG() types.ScriptCollection {
	collection := helpers.NewScriptCollection()
	collection.Append(types.Script{
		Name:           "k3s kubeconfig",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo cat /etc/rancher/k3s/k3s.yaml
`,
	})

	return collection
}

func (k3s *K3s) K8sVersion(ver string) types.KubernetesBootstrap {
	if err := isValidK3sVersion(ver); err == nil {
		// valid
		k3s.K3sVer = fmt.Sprintf("v%s+k3s1", ver)
		log.Debug(k3sCtx, "Printing", "k3s.K3sVer", k3s.K3sVer)
		return k3s
	} else {
		log.Error(k3sCtx, err.Error())
		return nil
	}
}

func (k3s *K3s) CNI(cni string) (externalCNI bool) {
	log.Debug(k3sCtx, "Printing", "cni", cni)
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
