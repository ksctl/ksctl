package kubernetes

var (
	apps map[string]func(string) Application
)

const (
	ArgocdStandardStackID         string = "standard-argocd"
	ArgoRolloutsStandardStackID   string = "standard-argorollouts"
	CiliumStandardStackID         string = "standard-cilium"
	FlannelStandardStackID        string = "standard-flannel"
	IstioStandardStackID          string = "standard-istio"
	KubePrometheusStandardStackID string = "standard-kubeprometheus"

	ArgocdProductionStackID   string = "production-argocd"
	KubeSpinProductionStackID string = "production-kubespin"
)

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)
