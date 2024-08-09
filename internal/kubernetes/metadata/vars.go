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
	SpinKubeProductionStackID StackID = "production-spinkube"
)

const (
	ArgocdComponentID       StackComponentID = "argocd"
	ArgorolloutsComponentID StackComponentID = "argorollouts"

	CiliumComponentID  StackComponentID = "cilium"
	FlannelComponentID StackComponentID = "flannel"

	CertManagerComponentID StackComponentID = "cert-manager"

	IstioComponentID StackComponentID = "istio"

	KubePrometheusComponentID StackComponentID = "kube-prometheus"

	KsctlApplicationComponentID StackComponentID = "ksctl-application-operator"

	KwasmOperatorComponentID StackComponentID = "kwasm-operator"

	SpinkubeOperatorCrdComponentID StackComponentID = "spinkube-operator-crd"
	SpinKubeOperatorRuntimeClassID StackComponentID = "spinkube-operator-runtime-class"
	SpinKubeOperatorShimExecutorID StackComponentID = "spinkube-operator-shim-executor"
	SpinKubeOperatorComponentID    StackComponentID = "spinkube-operator"
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)
