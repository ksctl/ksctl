package metadata

const (
	ArgocdStandardStackID         StackID = "standard-argocd"
	ArgoRolloutsStandardStackID   StackID = "standard-argorollouts"
	CiliumStandardStackID         StackID = "cilium"
	FlannelStandardStackID        StackID = "flannel"
	IstioStandardStackID          StackID = "standard-istio"
	KubePrometheusStandardStackID StackID = "standard-kubeprometheus"
	KsctlOperatorsID              StackID = "standard-ksctloperator"

	ArgocdProductionStackID   StackID = "production-argocd"
	KubeSpinProductionStackID StackID = "production-kubespin"
)

const (
	ArgocdComponentID           StackComponentID = "argocd"
	CertManagerComponentID      StackComponentID = "cert-manager"
	ArgorolloutsComponentID     StackComponentID = "argorollouts"
	CiliumComponentID           StackComponentID = "cilium"
	IstioComponentID            StackComponentID = "istio"
	FlannelComponentID          StackComponentID = "flannel"
	KubePrometheusComponentID   StackComponentID = "kube-prometheus"
	KsctlApplicationComponentID StackComponentID = "ksctl-application-operator"
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)
