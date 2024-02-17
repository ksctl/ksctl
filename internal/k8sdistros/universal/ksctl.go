package universal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ksctl/ksctl/internal/storage/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KSCTL_SYS_NAMESPACE = "ksctl"
)

var (

	// WARN: these rules only tested for the K3s distribution

	affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							corev1.NodeSelectorRequirement{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},

							corev1.NodeSelectorRequirement{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},
						},
					},
				},
			},
		},
	}

	tolerations = []corev1.Toleration{
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
	}
)

func (this *Kubernetes) DeleteResourcesFromController() error {
	log.Print("Started to configure Cluster to delete workerplanes")

	clusterName := this.Metadata.ClusterName
	region := this.Metadata.Region
	provider := this.Metadata.Provider
	distro := this.Metadata.K8sDistro

	log.Debug("Printing", "clustername", clusterName, "region", region, "provider", provider, "distro", distro)

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

			Affinity: affinity,

			Tolerations: tolerations,

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

	log.Debug("Printing", "destroyerPodManifest", destroyer)

	if err := this.PodApply(destroyer, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
	}

	count := consts.KsctlCounterConsts(0)
	for {

		status, err := this.clientset.CoreV1().Pods(KSCTL_SYS_NAMESPACE).Get(context.Background(), destroyer.ObjectMeta.Name, v1.GetOptions{})
		if err != nil {
			return log.NewError(err.Error())
		}
		if status.Status.Phase == corev1.PodSucceeded {
			log.Success(fmt.Sprintf("Status of Job [%v]", status.Status.Phase))
			break
		}
		count++
		if count == consts.CounterMaxRetryCount*2 {
			return log.NewError("max retry reached")
		}
		log.Debug(fmt.Sprintf("retrying current no of success [%v]", status.Status.Phase))
		time.Sleep(10 * time.Second)
	}

	time.Sleep(10 * time.Second) // to maintain a time gap for stable cluster and cloud resources

	log.Success("Done configuring Cluster to Scale down the no of workerplane to 1")
	return nil
}

func (this *Kubernetes) KsctlConfigForController(kubeconfig string, globalState *types.StorageDocument, secretKeys map[string][]byte) error {

	log.Print("Started to configure Cluster to add Ksctl specific resources")

	if err := this.namespaceCreate(KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
	}

	var globalStateRaw []byte
	var err error
	globalStateRaw, err = json.Marshal(globalState)
	if err != nil {
		return log.NewError(err.Error())
	}

	var state *corev1.ConfigMap = &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "ksctl-state",
		},
		BinaryData: map[string][]byte{
			"state.json": globalStateRaw,
		},
		//Data: map[string]any{
		//	"state.json": cloudstate,
		//},
	}

	log.Debug("Printing", "stateConfigMapManifest", state)

	if err := this.configMapApply(state, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
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

	log.Debug("Printing", "autoScalerControllerHelperConfigMapManifest", controllerInput)

	if err := this.configMapApply(controllerInput, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
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

	log.Debug("Printing", "cloudProviderSecretsManifest", tokenSecret)

	if err := this.secretApply(tokenSecret, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
	}

	// NOTE: reconstruct according to the http server
	newPath := fmt.Sprintf("/app/ksctl-data/.ksctl/state/%s/ha/%s", this.Metadata.Provider, this.Metadata.ClusterName+" "+this.Metadata.Region)

	replicas := int32(1)

	execNewPath := strings.Join(strings.Split(newPath, " "), "\\ ")
	rootFolder := fmt.Sprintf("/app/ksctl-data/.ksctl/config/%s/ha", this.Metadata.Provider)

	log.Debug("Printing", "newPathForClusterAccordingToPodFileSystem", newPath, "rootFolderForClusterAccordingToPodFileSystem", rootFolder)

	userid := int64(1000)
	groupid := int64(1000)
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
					Tolerations:       tolerations,
					PriorityClassName: "system-cluster-critical", // WARN: not sure if its okay

					Affinity: affinity,

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
								fmt.Sprintf("ls -la /tmp%s && mkdir -p %s && cp -v /tmp%s/..data/state.json %s/state.json", execNewPath, rootFolder, execNewPath, execNewPath),
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

	var listEnv []string

	switch this.Metadata.Provider {
	case consts.CloudCivo:
		listEnv = append(listEnv, "CIVO_TOKEN")
	case consts.CloudAzure:
		listEnv = append(listEnv, "AZURE_TENANT_ID", "AZURE_SUBSCRIPTION_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET")
	}

	deploymentEnv := make([]corev1.EnvVar, len(listEnv))

	for i, env := range listEnv {
		deploymentEnv[i] = corev1.EnvVar{
			Name: env,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: env,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: tokenSecret.ObjectMeta.Name,
					},
				},
			},
		}
	}
	ksctlServer.Spec.Template.Spec.Containers[0].Env = deploymentEnv

	log.Debug("Printing", "ksctlMainServerManifest", ksctlServer)

	time.Sleep(10 * time.Second) // waiting till the cluster is stable

	if err := this.deploymentApply(ksctlServer, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
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

	log.Debug("Printing", "ksctlServerServiceManifest", serverService)

	if err := this.serviceApply(serverService, KSCTL_SYS_NAMESPACE); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Done configuring Cluster to add Ksctl specific resources")
	return nil

}
