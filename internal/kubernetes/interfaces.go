package kubernetes

import (
	"context"

	"github.com/ksctl/ksctl/pkg/types"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	appsv1 "k8s.io/api/apps/v1"
)

type K8sClient interface {
	ApiExtensionsApply(ctx context.Context, log types.LoggerFactory, o *apiextensionsv1.CustomResourceDefinition) error
	ApiExtensionsDelete(ctx context.Context, log types.LoggerFactory, o *apiextensionsv1.CustomResourceDefinition) error
	ClusterRoleApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRole) error
	ClusterRoleBindingApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRoleBinding) error
	ClusterRoleBindingDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRoleBinding) error
	ClusterRoleDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRole) error
	ConfigMapApply(ctx context.Context, log types.LoggerFactory, o *corev1.ConfigMap) error
	ConfigMapDelete(ctx context.Context, log types.LoggerFactory, o *corev1.ConfigMap) error
	DaemonsetApply(ctx context.Context, log types.LoggerFactory, o *appsv1.DaemonSet) error
	DaemonsetDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.DaemonSet) error
	DeploymentApply(ctx context.Context, log types.LoggerFactory, o *appsv1.Deployment) error
	DeploymentDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.Deployment) error
	DeploymentReadyWait(ctx context.Context, log types.LoggerFactory, name string, namespace string) error
	JobApply(ctx context.Context, log types.LoggerFactory, o *batchv1.Job) error
	JobDelete(ctx context.Context, log types.LoggerFactory, o *batchv1.Job) error
	NamespaceCreate(ctx context.Context, log types.LoggerFactory, ns *corev1.Namespace) error
	NamespaceDelete(ctx context.Context, log types.LoggerFactory, ns *corev1.Namespace, wait bool) error
	NetPolicyApply(ctx context.Context, log types.LoggerFactory, o *netv1.NetworkPolicy) error
	NetPolicyDelete(ctx context.Context, log types.LoggerFactory, o *netv1.NetworkPolicy) error
	NodeDelete(ctx context.Context, log types.LoggerFactory, nodeName string) error
	NodesList(ctx context.Context, log types.LoggerFactory) (*corev1.NodeList, error)
	NodeUpdate(ctx context.Context, log types.LoggerFactory, node *corev1.Node) (*corev1.Node, error)
	PodApply(ctx context.Context, log types.LoggerFactory, o *corev1.Pod) error
	PodDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Pod) error
	PodReadyWait(ctx context.Context, log types.LoggerFactory, name string, namespace string) error
	RoleApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.Role) error
	RoleBindingApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.RoleBinding) error
	RoleBindingDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.RoleBinding) error
	RoleDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.Role) error
	SecretApply(ctx context.Context, log types.LoggerFactory, o *corev1.Secret) error
	SecretDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Secret) error
	ServiceAccountApply(ctx context.Context, log types.LoggerFactory, o *corev1.ServiceAccount) error
	ServiceAccountDelete(ctx context.Context, log types.LoggerFactory, o *corev1.ServiceAccount) error
	ServiceApply(ctx context.Context, log types.LoggerFactory, o *corev1.Service) error
	ServiceDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Service) error
	StatefulSetApply(ctx context.Context, log types.LoggerFactory, o *appsv1.StatefulSet) error
	StatefulSetDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.StatefulSet) error
	RuntimeApply(ctx context.Context, log types.LoggerFactory, o *nodev1.RuntimeClass) error
	RuntimeDelete(ctx context.Context, log types.LoggerFactory, resName string) error
}

type HelmClient interface {
	InstallChart(chartRef, chartVer, chartName, namespace, releaseName string, createNamespace bool, arguments map[string]interface{}) error
	ListInstalledCharts() error
	RepoAdd(repoName, repoUrl string) error
	UninstallChart(namespace, releaseName string) error
}
