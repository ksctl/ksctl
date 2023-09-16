package universal

import (
	"encoding/json"
	"fmt"

	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/utils"
	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) KsctlConfigForController(kubeconfig, kubeconfigpath, cloudstate, k8sstate string, secretKeys map[string][]byte) error {
	rawCloudstate := []byte(cloudstate)

	var sshPrivateKeyPath string
	var sshPubKeyPath string
	var clusterDir string

	switch this.Metadata.Provider {
	case utils.CLOUD_CIVO:
		var data *civo.StateConfiguration
		if err := json.Unmarshal(rawCloudstate, &data); err != nil {
			return err
		}
		sshPrivateKeyPath = data.SSHPrivateKeyLoc
		clusterDir = data.ClusterName + " " + data.Region

	case utils.CLOUD_AZURE:
		var data *azure.StateConfiguration
		if err := json.Unmarshal(rawCloudstate, &data); err != nil {
			return err
		}

		sshPrivateKeyPath = data.SSHPrivateKeyLoc
		clusterDir = data.ClusterName + " " + data.ResourceGroupName + " " + data.Region
	}

	sshPubKeyPath = sshPrivateKeyPath + ".pub"

	sshPrivate, err := this.StorageDriver.Path(sshPrivateKeyPath).Load()
	if err != nil {
		return err
	}

	sshPub, err := this.StorageDriver.Path(sshPubKeyPath).Load()
	if err != nil {
		return err
	}

	// reconstruct according to the http server
	newPath := fmt.Sprintf("/app/ksctl-data/config/%s/ha/%s", this.Metadata.Provider, clusterDir)

	fmt.Println("path for container -> ", newPath)

	replicas := int32(1)

	var ksctlServer *appsv1.Deployment = &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "ksctl-server",
			Labels: map[string]string{
				"app": "ksctl",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "ksctl",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": "ksctl",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:            "main",
							Image:           "docker.io/dipugodocker/kubesimplify:ksctl-slim-v1",
							ImagePullPolicy: corev1.PullAlways,

							EnvFrom: []corev1.EnvFromSource{
								corev1.EnvFromSource{
									Prefix: "CIVO_TOKEN",
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "cloud-secret",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "state-files",
									MountPath: newPath,
								},
								// corev1.VolumeMount{
								// 	Name:      "state-files-cloud",
								// 	MountPath: newPath + "/cloud-state.json",
								// },
								// corev1.VolumeMount{
								// 	Name:      "state-files-k8s",
								// 	MountPath: newPath + "/k8s-state.json",
								// },
								// corev1.VolumeMount{
								// 	Name:      "state-files-sshpub",
								// 	MountPath: newPath + "/keypair.pub",
								// },
								// corev1.VolumeMount{
								// 	Name:      "state-files-sshprivate",
								// 	MountPath: newPath + "/keypair",
								// },
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "server",
									ContainerPort: 80,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "state-files",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "ksctl-state",
									},
								},
							},
						},
					},
				},
			},
		},
	}

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
			"keypair.pub":      string(sshPub),
			"keypair":          string(sshPrivate),
		},
	}

	if err := this.configMapApply(state, "default"); err != nil {
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
			"VM_SIZE":      this.Metadata.WorkerPlaneNodeType,
			"NO_CP":        fmt.Sprint(this.Metadata.NoCP),
			"NO_WP":        fmt.Sprint(this.Metadata.NoWP),
		},
	}

	if err := this.configMapApply(controllerInput, "default"); err != nil {
		return err
	}

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

	if err := this.secretApply(tokenSecret, "default"); err != nil {
		return err
	}

	if err := this.deploymentApply(ksctlServer, "default"); err != nil {
		return err
	}
	return nil

}
