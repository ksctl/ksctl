package universal

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) KsctlConfigMap(kubeconfig string, kubeconfigpath string, cloudstate, k8sstate string) error {
	var state *corev1.ConfigMap = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "ksctl-state",
		},
		Data: map[string]string{
			"cloud-state.json": cloudstate,
			"k8s-state.json":   k8sstate,
			"kubeconfig":       kubeconfig,
			"kubeconfigpath":   kubeconfigpath,
		},
	}

	if err := this.ConfigMapApply(state, "default"); err != nil {
		return err
	}

	var controllerInput *corev1.ConfigMap = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "ksctl-controller",
		},
		Data: map[string]string{
			"CLUSTER_NAME": this.Metadata.ClusterName,
			"REGION":       this.Metadata.Region,
			"CLOUD":        this.Metadata.Provider,
			"DISTRO":       this.Metadata.K8sDistro,
			"K8S_VER":      this.Metadata.K8sVersion,
			"NO_CP":        fmt.Sprint(this.Metadata.NoCP),
			"NO_WP":        fmt.Sprint(this.Metadata.NoWP),
		},
	}

	if err := this.ConfigMapApply(controllerInput, "default"); err != nil {
		return err
	}

	// secret create
	// NOTE: get it from the cloud providers as map[string]string
	var secretKeys map[string][]byte // the tokens has to be in base64 form

	var tokenSecret *corev1.Secret = &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "cloud-secret",
		},
		Data: secretKeys,
	}
	fmt.Println(tokenSecret)
	return nil

}
