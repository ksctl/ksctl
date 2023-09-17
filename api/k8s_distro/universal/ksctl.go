package universal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/utils"
	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (this *Kubernetes) DeleteResourcesFromController() error {
	clusterName := this.Metadata.ClusterName
	region := this.Metadata.Region
	provider := this.Metadata.Provider
	distro := this.Metadata.K8sDistro
	var destroyer *corev1.Pod = &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "scale-to-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:    "destroyer",
					Image:   "alpine",
					Command: []string{"sh", "-c"},
					Args:    []string{fmt.Sprintf("apk add curl && curl -X PUT ksctl-server:8080/scaledown -d '{\"nowp\": 0, \"clustername\": \"%s\", \"region\": \"%s\", \"cloud\": \"%s\", \"distro\": \"%s\", \"vmsize\": \" \"}'", clusterName, region, provider, distro)},
					// TODO: add restart policy as never
				},
			},
		},
	}

	if err := this.PodApply(destroyer, "default"); err != nil {
		return err
	}

	// make the status check better as to when resume deletion of rest of the components
	status, err := this.clientset.CoreV1().Pods("default").Get(context.Background(), "scale-to-0", v1.GetOptions{})
	if err != nil {
		return err
	}
	fmt.Println(status.String())
	// wait for the resources to be destroyed
	time.Sleep(1 * time.Minute)

	return nil
}

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
		data.SSHPrivateKeyLoc = fmt.Sprintf("/app/ksctl-data/config/civo/ha/%s/keypair", clusterDir)
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		cloudstate = string(raw)

	case utils.CLOUD_AZURE:
		var data *azure.StateConfiguration
		if err := json.Unmarshal(rawCloudstate, &data); err != nil {
			return err
		}

		sshPrivateKeyPath = data.SSHPrivateKeyLoc
		clusterDir = data.ClusterName + " " + data.ResourceGroupName + " " + data.Region

		data.SSHPrivateKeyLoc = fmt.Sprintf("/app/ksctl-data/config/azure/ha/%s/keypair", clusterDir)
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		cloudstate = string(raw)
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

	// reconstruct according to the http server
	newPath := fmt.Sprintf("/app/ksctl-data/config/%s/ha/%s", this.Metadata.Provider, clusterDir)

	replicas := int32(1)

	execNewPath := strings.Join(strings.Split(newPath, " "), "\\ ")
	rootFolder := fmt.Sprintf("/app/ksctl-data/config/%s/ha", this.Metadata.Provider)

	// make it cloud provider specific
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
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:    "mount-points",
							Image:   "alpine",
							Command: []string{"sh", "-c"},
							Args: []string{
								fmt.Sprintf("ls -la /tmp%s && mkdir -p %s && cp -v /tmp%s/..data/kubeconfig %s/kubeconfig && cp -v /tmp%s/..data/cloud-state.json %s/cloud-state.json && cp -v /tmp%s/..data/k8s-state.json %s/k8s-state.json && cp -v /tmp%s/..data/keypair %s/keypair && cp -v /tmp%s/..data/keypair.pub %s/keypair.pub", execNewPath, rootFolder, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath, execNewPath),
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "cm-state-files",
									MountPath: fmt.Sprintf("/tmp%s", newPath),
								},

								corev1.VolumeMount{
									Name:      "main-data",
									MountPath: newPath,
								},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:            "main",
							Image:           "docker.io/dipugodocker/kubesimplify:ksctl-slim-v1",
							ImagePullPolicy: corev1.PullAlways,

							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "CIVO_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "CIVO_TOKEN",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "cloud-secret",
											},
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "main-data",
									MountPath: newPath,
								},
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
							Name: "cm-state-files",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "ksctl-state",
									},
								},
							},
						},
						corev1.Volume{
							Name: "main-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if err := this.deploymentApply(ksctlServer, "default"); err != nil {
		return err
	}

	var serverService *corev1.Service = &corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "ksctl-service",
			Labels: map[string]string{
				"app": "ksctl",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "ksctl",
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "hello",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	if err := this.serviceApply(serverService, "default"); err != nil {
		return err
	}

	return nil

}