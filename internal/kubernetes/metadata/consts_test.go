package metadata

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestStackIDConstants(t *testing.T) {
	assert.DeepEqual(t, "standard-argocd", string(ArgocdStandardStackID))
	assert.DeepEqual(t, "standard-argorollouts", string(ArgoRolloutsStandardStackID))
	assert.DeepEqual(t, "cilium", string(CiliumStandardStackID))
	assert.DeepEqual(t, "flannel", string(FlannelStandardStackID))
	assert.DeepEqual(t, "standard-istio", string(IstioStandardStackID))
	assert.DeepEqual(t, "standard-kubeprometheus", string(KubePrometheusStandardStackID))
	assert.DeepEqual(t, "standard-ksctloperator", string(KsctlOperatorsID))
	assert.DeepEqual(t, "production-argocd", string(ArgocdProductionStackID))
	assert.DeepEqual(t, "production-kubespin", string(KubeSpinProductionStackID))
}

func TestStackComponentIDConstants(t *testing.T) {
	assert.DeepEqual(t, "argocd", string(ArgocdComponentID))
	assert.DeepEqual(t, "cert-manager", string(CertManagerComponentID))
	assert.DeepEqual(t, "argorollouts", string(ArgorolloutsComponentID))
	assert.DeepEqual(t, "cilium", string(CiliumComponentID))
	assert.DeepEqual(t, "istio", string(IstioComponentID))
	assert.DeepEqual(t, "flannel", string(FlannelComponentID))
	assert.DeepEqual(t, "kube-prometheus", string(KubePrometheusComponentID))
	assert.DeepEqual(t, "ksctl-application-operator", string(KsctlApplicationComponentID))
}

func TestStackComponentTypeConstants(t *testing.T) {
	assert.DeepEqual(t, 0, int(ComponentTypeHelm))
	assert.DeepEqual(t, 1, int(ComponentTypeKubectl))
}
