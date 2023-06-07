package k8sdistro

type k8sDistro interface {
	GetServerToken() (string, error)
	GetKubeconfig() (string, error)
	SetupDatastore() (string, error)
	InitializeMasterControlPlane() error
	SetupWorkerplane() (string, error)
	JoinWorkerplane() (string, error)
	JoinControlplane() (string, error)
	JoinDatastore() (string, error)
}
