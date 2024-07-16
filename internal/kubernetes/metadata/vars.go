package metadata

const (
	ArgocdStandardStackID         StackID = "standard-argocd"
	ArgoRolloutsStandardStackID   StackID = "standard-argorollouts"
	CiliumStandardStackID         StackID = "standard-cilium"
	FlannelStandardStackID        StackID = "standard-flannel"
	IstioStandardStackID          StackID = "standard-istio"
	KubePrometheusStandardStackID StackID = "standard-kubeprometheus"
	KsctlApplicationOperatorID    StackID = "standard-ksctlapplicationoperator"

	ArgocdProductionStackID   StackID = "production-argocd"
	KubeSpinProductionStackID StackID = "production-kubespin"
)

const (
	ArgocdComponentID       StackComponentID = "argocd"
	CertManagerComponentID  StackComponentID = "cert-manager"
	ArgorolloutsComponentID StackComponentID = "argorollouts"
	RootComponentID         StackComponentID = "root"
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)
