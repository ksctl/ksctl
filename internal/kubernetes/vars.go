package kubernetes

var (
	apps map[string]func(string) Application
)

const (
	ArgocdStandardStackID   string = "standard-argocd"
	ArgocdProductionStackID string = "production-argocd"
)

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)
