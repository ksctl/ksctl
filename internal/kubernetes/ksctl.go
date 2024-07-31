package kubernetes

import (
	"fmt"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/ksctl-components/manifests"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KsctlSysNamespace      = "ksctl"
	KsctlServiceAcc        = "ksctl-sa"
	KsctlAgentClusterRole  = "ksctl-agent-crole"
	KsctlAgentCrbinding    = "ksctl-agent-croleb"
	KsctlAgentName         = "ksctl-agent"
	KsctlAgentService      = "agent"
	KsctlStateImporterName = "ksctl-state-importer"
	KsctlExtStoreSecret    = "ksctl-ext-store"
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

func (k *K8sClusterClient) DeployRequiredControllers(state *storageTypes.StorageDocument) error {
	log.Print(kubernetesCtx, "Started adding kubernetes ksctl specific controllers")
	apps := types.KsctlApp{
		StackName: string(metadata.KsctlOperatorsID),
		Overrides: map[string]map[string]any{
			string(metadata.KsctlApplicationComponentID): {
				"version": manifests.KsctlApplicationStackBranchOrTagName,
			},
		},
	}

	err := k.InstallApplication(apps, App, state)
	if err != nil {
		return err
	}

	log.Success(kubernetesCtx, "Done adding kubernetes ksctl specific controllers")
	return nil
}

func (k *K8sClusterClient) DeployAgent(client *types.KsctlClient,
	state *storageTypes.StorageDocument,
	externalStoreEndpoint map[string][]byte,
	v *types.StorageStateExportImport,
	isExternalStore bool) error {

	log.Print(kubernetesCtx, "Started to configure Cluster to add Ksctl specific storage")

	log.Print(kubernetesCtx, "creating ksctl namespace", "name", KsctlSysNamespace)
	if err := k.k8sClient.NamespaceCreate(
		kubernetesCtx,
		log,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: KsctlSysNamespace},
		}); err != nil {
		return err
	}

	serviceAccConfig := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KsctlServiceAcc,
			Namespace: KsctlSysNamespace,
			Labels:    labelsForKsctl,
		},
	}
	log.Print(kubernetesCtx, "creating service account for ksctl agent", "name", serviceAccConfig.Name)
	if err := k.k8sClient.ServiceAccountApply(
		kubernetesCtx,
		log,
		serviceAccConfig,
	); err != nil {
		return err
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   KsctlAgentClusterRole,
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
	if err := k.k8sClient.ClusterRoleApply(
		kubernetesCtx,
		log,
		clusterRole); err != nil {
		return err
	}

	clusterRoleBind := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   KsctlAgentCrbinding,
			Labels: labelsForKsctl,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccConfig.ObjectMeta.Name,
				Namespace: KsctlSysNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     clusterRole.ObjectMeta.Name,
			Kind:     "ClusterRole",
		},
	}

	log.Print(kubernetesCtx, "creating clusterrolebinding", "name", clusterRoleBind.Name)
	if err := k.k8sClient.ClusterRoleBindingApply(
		kubernetesCtx,
		log,
		clusterRoleBind); err != nil {
		return err
	}

	if !isExternalStore {

		var ksctlStateImporter *corev1.Pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KsctlStateImporterName,
				Namespace: KsctlSysNamespace,
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
									Port: intstr.FromInt32(8080),
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
		if err := k.k8sClient.PodApply(
			kubernetesCtx,
			log,
			ksctlStateImporter); err != nil {
			return err
		}

		if err := k.k8sClient.PodReadyWait(
			kubernetesCtx,
			log,
			ksctlStateImporter.Name, ksctlStateImporter.Namespace); err != nil {
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
		if err := k.k8sClient.PodDelete(
			kubernetesCtx,
			log,
			ksctlStateImporter); err != nil {
			return err
		}
	}

	replicas := int32(1)

	agentSelector := utilities.DeepCopyMap[string, string](labelsForKsctl)
	agentSelector["scope"] = "agent"
	var ksctlServer *appsv1.Deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KsctlAgentName,
			Labels:    agentSelector,
			Namespace: KsctlSysNamespace,
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
				Name:      KsctlExtStoreSecret,
				Labels:    labelsForKsctl,
				Namespace: KsctlSysNamespace,
			},
			Data: externalStoreEndpoint,
		}

		log.Print(kubernetesCtx, "creating external store secrets for ksctl agent", "name", secretExt.Name)

		if err := k.k8sClient.SecretApply(
			kubernetesCtx,
			log,
			secretExt); err != nil {
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
	if err := k.k8sClient.DeploymentApply(
		kubernetesCtx,
		log,
		ksctlServer); err != nil {
		return err
	}

	var serverService *corev1.Service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KsctlAgentService,
			Labels:    labelsForKsctl,
			Namespace: KsctlSysNamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: agentSelector,

			Type: corev1.ServiceTypeClusterIP,

			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	log.Print(kubernetesCtx, "creating ksctl agent service", "name", serverService.Name)
	if err := k.k8sClient.ServiceApply(
		kubernetesCtx,
		log,
		serverService); err != nil {
		return err
	}

	if err := k.k8sClient.DeploymentReadyWait(
		kubernetesCtx,
		log,
		ksctlServer.Name, ksctlServer.Namespace); err != nil {
		return err
	}

	log.Success(kubernetesCtx, "Done configuring Cluster to add Ksctl specific storage")
	return nil

}
