package k3s

import (
	"fmt"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"sync"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
)

type K3s struct {
	K3sVer string
	Cni    string
	mu     *sync.Mutex
}

func NewClient(m types.Metadata, state *storageTypes.StorageDocument) types.KubernetesBootstrap {
	log = logger.NewStructuredLogger(m.LogVerbosity, m.LogWritter)
	log.SetPackageName("k3s")

	mainStateDocument = state
	return &K3s{mu: &sync.Mutex{}}
}

func (k3s *K3s) Setup(storage types.StorageFactory, operation consts.KsctlOperation) error {
	if operation == consts.OperationCreate {
		mainStateDocument.K8sBootstrap.K3s = &storageTypes.StateConfigurationK3s{}
		mainStateDocument.BootstrapProvider = consts.K8sK3s
	}

	if err := storage.Write(mainStateDocument); err != nil {
		return log.NewError(err.Error())
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

func (k3s *K3s) Version(ver string) types.KubernetesBootstrap {
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
