package universal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	corev1 "k8s.io/api/core/v1"

	. "github.com/kubesimplify/ksctl/api/utils/consts"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	KSCTL_SYS_NAMESPACE = "ksctl"
)

func (this *Kubernetes) DeleteResourcesFromController() error {
	this.StorageDriver.Logger().Print("[client-go] Started to configure Cluster to delete workerplanes")
	clusterName := this.Metadata.ClusterName
	region := this.Metadata.Region
	provider := this.Metadata.Provider
	distro := this.Metadata.K8sDistro

	// TODO: make the node have toleration for CriticalAddonsOnly=true:NoExecute to schedule in one of the controlplane

	var destroyer *corev1.Pod = &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "scale-to-0",
		},
		Spec: corev1.PodSpec{

			PriorityClassName: "system-cluster-critical", // WARN: not sure if its okay

			// WARN: these toleration rules only tested for the K3s distribution
			Tolerations: []corev1.Toleration{
				corev1.Toleration{
					Key:      "CriticalAddonsOnly",
					Operator: corev1.TolerationOpExists,
				},
				corev1.Toleration{
					Effect:   corev1.TaintEffectNoSchedule,
					Key:      "node-role.kubernetes.io/control-plane",
					Operator: corev1.TolerationOpExists,
				},
				corev1.Toleration{
					Effect:   corev1.TaintEffectNoSchedule,
					Key:      "node-role.kubernetes.io/master",
					Operator: corev1.TolerationOpExists,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:    "destroyer",
					Image:   "alpine",
					Command: []string{"sh", "-c"},
					Args:    []string{fmt.Sprintf("apk add curl && curl -X PUT ksctl-service:8080/scaledown -d '{\"nowp\": 0, \"clustername\": \"%s\", \"region\": \"%s\", \"cloud\": \"%s\", \"distro\": \"%s\"}'", clusterName, region, provider, distro)},

					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("10m"),
							corev1.ResourceMemory: resource.MustParse("10Mi"),
						},
					},
				},
			},
		},
	}

	if err := this.PodApply(destroyer, KSCTL_SYS_NAMESPACE); err != nil {
		return err
	}

	count := KsctlCounterConts(0)
	for {

		status, err := this.clientset.CoreV1().Pods(KSCTL_SYS_NAMESPACE).Get(context.Background(), destroyer.ObjectMeta.Name, v1.GetOptions{})
		if err != nil {
			return err
		}
		if status.Status.Phase == corev1.PodSucceeded {
			this.StorageDriver.Logger().Success(fmt.Sprintf("Status of Job [%v]", status.Status.Phase))
			break
		}
		count++
		if count == MAX_RETRY_COUNT {
			return fmt.Errorf("max retry reached")
		}
		this.StorageDriver.Logger().Warn(fmt.Sprintf("retrying current no of success [%v]", status.Status.Phase))
		time.Sleep(10 * time.Second)
	}

	time.Sleep(10 * time.Second) // to maintain a time gap for stable cluster and cloud resources

	this.StorageDriver.Logger().Success("[client-go] Done configuring Cluster to Scale down the no of workerplane to 1")
	return nil
}

func (this *Kubernetes) KsctlConfigForController(kubeconfig, kubeconfigpath, cloudstate, k8sstate string, secretKeys map[string][]byte) error {

	this.StorageDriver.Logger().Print("[client-go] Started to configure Cluster to add Ksctl specific resources")
	rawCloudstate := []byte(cloudstate)

	var sshPrivateKeyPath string
	var sshPubKeyPath string
	var clusterDir string

	if err := this.namespaceCreate(KSCTL_SYS_NAMESPACE); err != nil {
		return err
	}

	switch this.Metadata.Provider {
	case CLOUD_CIVO:
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

	case CLOUD_AZURE:
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

	if err := this.configMapApply(state, KSCTL_SYS_NAMESPACE); err != nil {
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
			"CLOUD":        string(this.Metadata.Provider),
			"DISTRO":       string(this.Metadata.K8sDistro),
			"K8S_VER":      this.Metadata.K8sVersion,
			"VM_SIZE":      this.Metadata.WorkerPlaneNodeType,
			"NO_CP":        fmt.Sprint(this.Metadata.NoCP),
			"NO_WP":        fmt.Sprint(this.Metadata.NoWP),
		},
	}

	if err := this.configMapApply(controllerInput, KSCTL_SYS_NAMESPACE); err != nil {
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

	if err := this.secretApply(tokenSecret, KSCTL_SYS_NAMESPACE); err != nil {
		return err
	}

	// reconstruct according to the http server
	newPath := fmt.Sprintf("/app/ksctl-data/config/%s/ha/%s", this.Metadata.Provider, clusterDir)

	replicas := int32(1)

	execNewPath := strings.Join(strings.Split(newPath, " "), "\\ ")
	rootFolder := fmt.Sprintf("/app/ksctl-data/config/%s/ha", this.Metadata.Provider)

	userid := int64(1000)
	groupid := int64(1000)
	// make it cloud provider specific
	// TODO: make the node have toleration for CriticalAddonsOnly=true:NoExecute to schedule in one of the controlplane
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
					// WARN: these toleration rules only tested for the K3s distribution
					Tolerations: []corev1.Toleration{
						corev1.Toleration{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
						corev1.Toleration{
							Effect:   corev1.TaintEffectNoSchedule,
							Key:      "node-role.kubernetes.io/control-plane",
							Operator: corev1.TolerationOpExists,
						},
						corev1.Toleration{
							Effect:   corev1.TaintEffectNoSchedule,
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
						},
					},
					PriorityClassName: "system-cluster-critical", // WARN: not sure if its okay

					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  &userid,
						RunAsGroup: &groupid,
					},
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
							Image:           "docker.io/kubesimplify/ksctl:slim-v1",
							ImagePullPolicy: corev1.PullAlways,
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
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("250Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
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
										Name: state.ObjectMeta.Name,
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

	switch this.Metadata.Provider {
	case CLOUD_CIVO:
		ksctlServer.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
			corev1.EnvVar{
				Name: "CIVO_TOKEN",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "CIVO_TOKEN",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tokenSecret.ObjectMeta.Name,
						},
					},
				},
			},
		}
	case CLOUD_AZURE:
		ksctlServer.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
			corev1.EnvVar{
				Name: "AZURE_TENANT_ID",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "AZURE_TENANT_ID",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tokenSecret.ObjectMeta.Name,
						},
					},
				},
			},
			corev1.EnvVar{
				Name: "AZURE_SUBSCRIPTION_ID",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "AZURE_SUBSCRIPTION_ID",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tokenSecret.ObjectMeta.Name,
						},
					},
				},
			},
			corev1.EnvVar{
				Name: "AZURE_CLIENT_ID",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "AZURE_CLIENT_ID",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tokenSecret.ObjectMeta.Name,
						},
					},
				},
			},
			corev1.EnvVar{
				Name: "AZURE_CLIENT_SECRET",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "AZURE_CLIENT_SECRET",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tokenSecret.ObjectMeta.Name,
						},
					},
				},
			},
		}
	}

	time.Sleep(10 * time.Second) // waiting till the cluster is stable

	if err := this.deploymentApply(ksctlServer, KSCTL_SYS_NAMESPACE); err != nil {
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

	if err := this.serviceApply(serverService, KSCTL_SYS_NAMESPACE); err != nil {
		return err
	}

	this.StorageDriver.Logger().Success("[client-go] Done configuring Cluster to add Ksctl specific resources")
	return nil

}
