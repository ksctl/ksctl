package kubernetes

var (
	apps map[string]func(string) Application
)

const (
	ArgocdStandardStackID     string = "standard-argocd"
	ArgocdProductionStackID   string = "production-argocd"
	KubeSpinProductionStackID string = "production-kubespin"
)

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)
