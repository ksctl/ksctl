package kubernetes

import (
	"fmt"

	"github.com/ksctl/ksctl/ksctl-components/manifests"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KSCTL_SYS_NAMESPACE       = "ksctl"
	KSCTL_SERVICE_ACC         = "ksctl-sa"
	KSCTL_AGENT_CLUSTER_ROLE  = "ksctl-agent-crole"
	KSCTL_AGENT_CRBINDING     = "ksctl-agent-croleb"
	KSCTL_AGENT_NAME          = "ksctl-agent"
	KSCTL_AGENT_SERVICE       = "agent"
	KSCTL_STATE_IMPORTER_NAME = "ksctl-state-importer"
	KSCTL_EXT_STORE_SECRET    = "ksctl-ext-store"
)

var (
	tolerations = []corev1.Toleration{
		{
			Key:      "CriticalAddonsOnly",
			Operator: corev1.TolerationOpExists,
		},
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Key:      "node-role.kubernetes.io/control-plane",
			Operator: corev1.TolerationOpExists,
		},
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Key:      "node-role.kubernetes.io/master",
			Operator: corev1.TolerationOpExists,
		},
	}

	labelsForKsctl map[string]string = map[string]string{
		"app.kubernetes.io/name":    "ksctl",
		"app.kubernetes.io/part-of": "ksctl-components",
	}
)

func (k *Kubernetes) DeployRequiredControllers(state *storageTypes.StorageDocument, isExternalStore bool) error {
	log.Print(kubernetesCtx, "Started adding kubernetes ksctl specific controllers")
	components := []string{KsctlApplicationOperatorID + "@" + manifests.KsctlApplicationStackBranchOrTagName}

	_apps, err := helpers.ToApplicationTempl(kubernetesCtx, log, components)
	if err != nil {
		return err
	}
	err = k.Applications(_apps, state, consts.OperationCreate)
	if err != nil {
		return err
	}

	log.Success(kubernetesCtx, "Done adding kubernetes ksctl specific controllers")
	return nil
}

func (k *Kubernetes) DeployAgent(client *types.KsctlClient,
	state *storageTypes.StorageDocument,
	externalStoreEndpoint map[string][]byte,
	v *types.StorageStateExportImport,
	isExternalStore bool) error {

	log.Print(kubernetesCtx, "Started to configure Cluster to add Ksctl specific storage")

	log.Print(kubernetesCtx, "creating ksctl namespace", "name", KSCTL_SYS_NAMESPACE)
	if err := k.namespaceCreate(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: KSCTL_SYS_NAMESPACE},
	}); err != nil {
		return err
	}

	serviceAccConfig := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSCTL_SERVICE_ACC,
			Namespace: KSCTL_SYS_NAMESPACE,
			Labels:    labelsForKsctl,
		},
	}
	log.Print(kubernetesCtx, "creating service account for ksctl agent", "name", serviceAccConfig.Name)
	if err := k.serviceAccountApply(serviceAccConfig); err != nil {
		return err
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   KSCTL_AGENT_CLUSTER_ROLE,
			Labels: labelsForKsctl,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	log.Print(kubernetesCtx, "creating clusterrole", "name", clusterRole.Name)
	if err := k.clusterRoleApply(clusterRole); err != nil {
		return err
	}

	clusterRoleBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   KSCTL_AGENT_CRBINDING,
			Labels: labelsForKsctl,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccConfig.ObjectMeta.Name,
				Namespace: KSCTL_SYS_NAMESPACE,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     clusterRole.ObjectMeta.Name,
			Kind:     "ClusterRole",
		},
	}

	log.Print(kubernetesCtx, "creating clusterrolebinding", "name", clusterRoleBind.Name)
	if err := k.clusterRoleBindingApply(clusterRoleBind); err != nil {
		return err
	}

	if !isExternalStore {

		var ksctlStateImporter *corev1.Pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KSCTL_STATE_IMPORTER_NAME,
				Namespace: KSCTL_SYS_NAMESPACE,
			},
			Spec: corev1.PodSpec{
				RestartPolicy:      corev1.RestartPolicyAlways,
				ServiceAccountName: serviceAccConfig.ObjectMeta.Name,
				Containers: []corev1.Container{
					{
						Name:            "ksctl-stateimport",
						Image:           "ghcr.io/ksctl/ksctl-stateimport:" + manifests.KsctlStateImportAppVersion,
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{
							{
								Name:          "service",
								ContainerPort: 80,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "LOG_LEVEL",
								Value: "DEBUG",
							},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromInt(8080),
								},
							},
							InitialDelaySeconds: 5,
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
			},
		}

		log.Print(kubernetesCtx, "creating ksctl state transfer pod", "name", ksctlStateImporter.Name)
		if err := k.PodApply(ksctlStateImporter); err != nil {
			return err
		}

		if err := k.podReadyWait(ksctlStateImporter.Name, ksctlStateImporter.Namespace); err != nil {
			return err
		}

		log.Print(kubernetesCtx, "transfer data by making http call")

		kubeconfig := state.ClusterKubeConfig
		kubeconfigContext := state.ClusterKubeConfigContext
		if err := transferData(
			kubeconfig,
			kubeconfigContext,
			ksctlStateImporter.Name,
			ksctlStateImporter.Namespace,
			8080,
			v,
		); err != nil {
			return err
		}

		log.Print(kubernetesCtx, "destroying the state importer", "name", ksctlStateImporter.Name)
		if err := k.PodDelete(ksctlStateImporter); err != nil {
			return err
		}
	}

	replicas := int32(1)

	agentSelector := utilities.DeepCopyMap[string, string](labelsForKsctl)
	agentSelector["scope"] = "agent"
	var ksctlServer *appsv1.Deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSCTL_AGENT_NAME,
			Labels:    agentSelector,
			Namespace: KSCTL_SYS_NAMESPACE,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: agentSelector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: agentSelector,
				},
				Spec: corev1.PodSpec{
					Tolerations:       tolerations,
					PriorityClassName: "system-cluster-critical", // WARN: not sure if its okay

					// Affinity:      affinity,
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: serviceAccConfig.ObjectMeta.Name,
					Containers: []corev1.Container{
						{
							Name:            "ksctl-agent",
							Image:           "ghcr.io/ksctl/ksctl-agent:" + manifests.KsctlAgentAppVersion,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc-server",
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "LOG_LEVEL",
									Value: "DEBUG",
								},
								{
									Name:  "KSCTL_CLUSTER_IS_HA",
									Value: fmt.Sprintf("%v", client.Metadata.IsHA),
								},
								{
									Name:  "KSCTL_CLUSTER_NAME",
									Value: client.Metadata.ClusterName,
								},
								{
									Name:  "KSCTL_CLOUD",
									Value: string(client.Metadata.Provider),
								},
								{
									Name:  "KSCTL_REGION",
									Value: client.Metadata.Region,
								},
								{
									Name:  "KSCTL_K8S_DISTRO",
									Value: string(client.Metadata.K8sDistro),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									GRPC: &corev1.GRPCAction{
										Port: 8080,
									},
								},
								InitialDelaySeconds: 5,
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
				},
			},
		},
	}

	if isExternalStore {

		secretExt := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KSCTL_EXT_STORE_SECRET,
				Labels:    labelsForKsctl,
				Namespace: KSCTL_SYS_NAMESPACE,
			},
			Data: externalStoreEndpoint,
		}

		log.Print(kubernetesCtx, "creating external store secrets for ksctl agent", "name", secretExt.Name)

		if err := k.secretApply(secretExt); err != nil {
			return err
		}

		for k, _ := range externalStoreEndpoint {
			ksctlServer.Spec.Template.Spec.Containers[0].Env = append(ksctlServer.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name: k,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: k,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretExt.ObjectMeta.Name,
						},
					},
				}})
		}

	}

	log.Print(kubernetesCtx, "creating ksctl agent deployment", "name", ksctlServer.Name)
	if err := k.deploymentApply(ksctlServer); err != nil {
		return err
	}

	var serverService *corev1.Service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSCTL_AGENT_SERVICE,
			Labels:    labelsForKsctl,
			Namespace: KSCTL_SYS_NAMESPACE,
		},
		Spec: corev1.ServiceSpec{
			Selector: agentSelector,

			Type: corev1.ServiceTypeClusterIP,

			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	log.Print(kubernetesCtx, "creating ksctl agent service", "name", serverService.Name)
	if err := k.serviceApply(serverService); err != nil {
		return err
	}

	if err := k.deploymentReadyWait(ksctlServer.Name, ksctlServer.Namespace); err != nil {
		return err
	}

	log.Success(kubernetesCtx, "Done configuring Cluster to add Ksctl specific storage")
	return nil

}
