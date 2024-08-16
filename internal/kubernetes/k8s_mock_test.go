package kubernetes

import (
	"context"

	"github.com/ksctl/ksctl/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8sClientMock struct{}

func (k k8sClientMock) RuntimeApply(ctx context.Context, log types.LoggerFactory, o *v1.RuntimeClass) error {
	return nil
}

func (k k8sClientMock) RuntimeDelete(ctx context.Context, log types.LoggerFactory, resName string) error {
	return nil
}

func (k k8sClientMock) NodeUpdate(ctx context.Context, log types.LoggerFactory, node *corev1.Node) (*corev1.Node, error) {
	return node, nil
}

func (k k8sClientMock) ApiExtensionsApply(ctx context.Context, log types.LoggerFactory, o *apiextensionsv1.CustomResourceDefinition) error {
	return nil
}

func (k k8sClientMock) ApiExtensionsDelete(ctx context.Context, log types.LoggerFactory, o *apiextensionsv1.CustomResourceDefinition) error {
	return nil
}

func (k k8sClientMock) ClusterRoleApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRole) error {
	return nil
}

func (k k8sClientMock) ClusterRoleBindingApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRoleBinding) error {
	return nil
}

func (k k8sClientMock) ClusterRoleBindingDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRoleBinding) error {
	return nil
}

func (k k8sClientMock) ClusterRoleDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.ClusterRole) error {
	return nil
}

func (k k8sClientMock) ConfigMapApply(ctx context.Context, log types.LoggerFactory, o *corev1.ConfigMap) error {
	return nil
}

func (k k8sClientMock) ConfigMapDelete(ctx context.Context, log types.LoggerFactory, o *corev1.ConfigMap) error {
	return nil
}

func (k k8sClientMock) DaemonsetApply(ctx context.Context, log types.LoggerFactory, o *appsv1.DaemonSet) error {
	return nil
}

func (k k8sClientMock) DaemonsetDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.DaemonSet) error {
	return nil
}

func (k k8sClientMock) DeploymentApply(ctx context.Context, log types.LoggerFactory, o *appsv1.Deployment) error {
	return nil
}

func (k k8sClientMock) DeploymentDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.Deployment) error {
	return nil
}

func (k k8sClientMock) DeploymentReadyWait(ctx context.Context, log types.LoggerFactory, name string, namespace string) error {
	return nil
}

func (k k8sClientMock) JobApply(ctx context.Context, log types.LoggerFactory, o *batchv1.Job) error {
	return nil
}

func (k k8sClientMock) JobDelete(ctx context.Context, log types.LoggerFactory, o *batchv1.Job) error {
	return nil
}

func (k k8sClientMock) NamespaceCreate(ctx context.Context, log types.LoggerFactory, ns *corev1.Namespace) error {
	return nil
}

func (k k8sClientMock) NamespaceDelete(ctx context.Context, log types.LoggerFactory, ns *corev1.Namespace, wait bool) error {
	return nil
}

func (k k8sClientMock) NetPolicyApply(ctx context.Context, log types.LoggerFactory, o *networkingv1.NetworkPolicy) error {
	return nil
}

func (k k8sClientMock) NetPolicyDelete(ctx context.Context, log types.LoggerFactory, o *networkingv1.NetworkPolicy) error {
	return nil
}

func (k k8sClientMock) NodeDelete(ctx context.Context, log types.LoggerFactory, nodeName string) error {
	return nil
}

func (k k8sClientMock) NodesList(ctx context.Context, log types.LoggerFactory) (*corev1.NodeList, error) {
	return &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
			},
		},
	}, nil
}

func (k k8sClientMock) PodApply(ctx context.Context, log types.LoggerFactory, o *corev1.Pod) error {
	return nil
}

func (k k8sClientMock) PodDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Pod) error {
	return nil
}

func (k k8sClientMock) PodReadyWait(ctx context.Context, log types.LoggerFactory, name string, namespace string) error {
	return nil
}

func (k k8sClientMock) RoleApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.Role) error {
	return nil
}

func (k k8sClientMock) RoleBindingApply(ctx context.Context, log types.LoggerFactory, o *rbacv1.RoleBinding) error {
	return nil
}

func (k k8sClientMock) RoleBindingDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.RoleBinding) error {
	return nil
}

func (k k8sClientMock) RoleDelete(ctx context.Context, log types.LoggerFactory, o *rbacv1.Role) error {
	return nil
}

func (k k8sClientMock) SecretApply(ctx context.Context, log types.LoggerFactory, o *corev1.Secret) error {
	return nil
}

func (k k8sClientMock) SecretDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Secret) error {
	return nil
}

func (k k8sClientMock) ServiceAccountApply(ctx context.Context, log types.LoggerFactory, o *corev1.ServiceAccount) error {
	return nil
}

func (k k8sClientMock) ServiceAccountDelete(ctx context.Context, log types.LoggerFactory, o *corev1.ServiceAccount) error {
	return nil
}

func (k k8sClientMock) ServiceApply(ctx context.Context, log types.LoggerFactory, o *corev1.Service) error {
	return nil
}

func (k k8sClientMock) ServiceDelete(ctx context.Context, log types.LoggerFactory, o *corev1.Service) error {
	return nil
}

func (k k8sClientMock) StatefulSetApply(ctx context.Context, log types.LoggerFactory, o *appsv1.StatefulSet) error {
	return nil
}

func (k k8sClientMock) StatefulSetDelete(ctx context.Context, log types.LoggerFactory, o *appsv1.StatefulSet) error {
	return nil
}
