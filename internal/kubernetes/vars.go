package kubernetes

var (
	apps map[string]func(string) Application

	appsManifest map[string]func(applicationParams) ApplicationStack = map[string]func(applicationParams) ApplicationStack{
		ArgocdStandardStackID:         argocdStandardCICD,
		ArgoRolloutsStandardStackID:   argoRolloutsStandardCICD,
		CiliumStandardStackID:         ciliumStandardCNI,
		FlannelStandardStackID:        flannelStandardCNI,
		IstioStandardStackID:          istioStandardServiceMesh,
		KubePrometheusStandardStackID: kubePrometheusStandardMonitoring,
	}
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
