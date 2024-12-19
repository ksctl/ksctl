package distributions

import "github.com/ksctl/ksctl/pkg/consts"

type KubernetesDistribution interface {
	Setup(storage Storage, operation consts.KsctlOperation) error

	ConfigureControlPlane(int, Storage) error

	JoinWorkerplane(int, Storage) error

	K8sVersion(string) KubernetesDistribution

	CNI(string) (externalCNI bool)
}
