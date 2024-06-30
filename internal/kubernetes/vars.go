package kubernetes

var (
	apps map[string]func(string) Application
)

const (
	ArgocdStandardStackID       string = "standard-argocd"
	ArgoRolloutsStandardStackID string = "standard-argorollouts"
	ArgocdProductionStackID     string = "production-argocd"
	KubeSpinProductionStackID   string = "production-kubespin"
)

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)
