package kubernetes

var (
	appsManifest map[string]func(applicationParams) ApplicationStack = map[string]func(applicationParams) ApplicationStack{
		ArgocdStandardStackID:         argocdStandardCICD,
		ArgoRolloutsStandardStackID:   argoRolloutsStandardCICD,
		CiliumStandardStackID:         ciliumStandardCNI,
		FlannelStandardStackID:        flannelStandardCNI,
		IstioStandardStackID:          istioStandardServiceMesh,
		KubePrometheusStandardStackID: kubePrometheusStandardMonitoring,
		KsctlApplicationOperatorID:    applicationStackData,
	}
)

const (
	ArgocdStandardStackID         string = "standard-argocd"
	ArgoRolloutsStandardStackID   string = "standard-argorollouts"
	CiliumStandardStackID         string = "standard-cilium"
	FlannelStandardStackID        string = "standard-flannel"
	IstioStandardStackID          string = "standard-istio"
	KubePrometheusStandardStackID string = "standard-kubeprometheus"
	KsctlApplicationOperatorID    string = "standard-ksctlapplicationoperator"

	ArgocdProductionStackID   string = "production-argocd"
	KubeSpinProductionStackID string = "production-kubespin"
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)
