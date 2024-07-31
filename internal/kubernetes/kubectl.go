package kubernetes

import (
	"io"
	"net/http"
	"strings"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func getManifests(appUrl string) ([]string, error) {

	resp, err := http.Get(appUrl)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to get manifests", "Reason", err),
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to ready manifests", "Reason", err),
		)
	}

	resources := strings.Split(string(body), "---")
	if err := apiextensionsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to add apiextensionv1 to the scheme", "Reason", err),
		)
	}
	return resources, nil
}

func deleteKubectl(client *K8sClusterClient, component *metadata.KubectlHandler) error {
	resources, err := getManifests(component.Url)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err),
			)
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = client.k8sClient.ApiExtensionsDelete(kubernetesCtx, log, o)

		case *corev1.Namespace:
			log.Print(kubernetesCtx, "Namespace", "name", o.Name)
			errRes = client.k8sClient.NamespaceDelete(kubernetesCtx, log, o, false)

		case *appsv1.DaemonSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Daemonset", "name", o.Name)
			errRes = client.k8sClient.DaemonsetDelete(kubernetesCtx, log, o)

		case *appsv1.Deployment:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Deployment", "name", o.Name)
			errRes = client.k8sClient.DeploymentDelete(kubernetesCtx, log, o)

		case *corev1.Service:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Service", "name", o.Name)
			errRes = client.k8sClient.ServiceDelete(kubernetesCtx, log, o)

		case *corev1.ServiceAccount:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ServiceAccount", "name", o.Name)
			errRes = client.k8sClient.ServiceAccountDelete(kubernetesCtx, log, o)

		case *corev1.ConfigMap:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ConfigMap", "name", o.Name)
			errRes = client.k8sClient.ConfigMapDelete(kubernetesCtx, log, o)

		case *corev1.Secret:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Secret", "name", o.Name)
			errRes = client.k8sClient.SecretDelete(kubernetesCtx, log, o)

		case *appsv1.StatefulSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "StatefulSet", "name", o.Name)
			errRes = client.k8sClient.StatefulSetDelete(kubernetesCtx, log, o)

		case *rbacv1.ClusterRole:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ClusterRole", "name", o.Name)
			errRes = client.k8sClient.ClusterRoleDelete(kubernetesCtx, log, o)

		case *rbacv1.ClusterRoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ClusterRoleBinding", "name", o.Name)
			errRes = client.k8sClient.ClusterRoleBindingDelete(kubernetesCtx, log, o)

		case *rbacv1.Role:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Role", "name", o.Name)
			errRes = client.k8sClient.RoleDelete(kubernetesCtx, log, o)

		case *rbacv1.RoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "RoleBinding", "name", o.Name)
			errRes = client.k8sClient.RoleBindingDelete(kubernetesCtx, log, o)

		case *networkingv1.NetworkPolicy:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "NetworkPolicy", "name", o.Name)
			errRes = client.k8sClient.NetPolicyDelete(kubernetesCtx, log, o)

		default:
			errRes = ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "unexpected type", "obj", o),
			)
		}

		if errRes != nil {
			return errRes
		}
	}

	if component.CreateNamespace {
		if err := client.k8sClient.NamespaceDelete(kubernetesCtx, log, &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: component.Namespace,
			}}, true); err != nil {
			return err
		}
	}

	return nil
}

func installKubectl(client *K8sClusterClient, component *metadata.KubectlHandler) error {
	resources, err := getManifests(component.Url)
	if err != nil {
		return err
	}

	if component.CreateNamespace {
		if err := client.k8sClient.NamespaceCreate(kubernetesCtx, log, &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: component.Namespace,
			}}); err != nil {
			return err
		}
	}

	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err),
			)
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = client.k8sClient.ApiExtensionsApply(kubernetesCtx, log, o)

		case *corev1.Namespace:
			log.Print(kubernetesCtx, "Namespace", "name", o.Name)
			errRes = client.k8sClient.NamespaceCreate(kubernetesCtx, log, o)

		case *appsv1.DaemonSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Daemonset", "name", o.Name)
			errRes = client.k8sClient.DaemonsetApply(kubernetesCtx, log, o)

		case *appsv1.Deployment:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Deployment", "name", o.Name)
			errRes = client.k8sClient.DeploymentApply(kubernetesCtx, log, o)

		case *corev1.Service:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Service", "name", o.Name)
			errRes = client.k8sClient.ServiceApply(kubernetesCtx, log, o)

		case *corev1.ServiceAccount:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ServiceAccount", "name", o.Name)
			errRes = client.k8sClient.ServiceAccountApply(kubernetesCtx, log, o)

		case *corev1.ConfigMap:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ConfigMap", "name", o.Name)
			errRes = client.k8sClient.ConfigMapApply(kubernetesCtx, log, o)

		case *corev1.Secret:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Secret", "name", o.Name)
			errRes = client.k8sClient.SecretApply(kubernetesCtx, log, o)

		case *appsv1.StatefulSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "StatefulSet", "name", o.Name)
			errRes = client.k8sClient.StatefulSetApply(kubernetesCtx, log, o)

		case *rbacv1.ClusterRole:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ClusterRole", "name", o.Name)
			errRes = client.k8sClient.ClusterRoleApply(kubernetesCtx, log, o)

		case *rbacv1.ClusterRoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "ClusterRoleBinding", "name", o.Name)
			errRes = client.k8sClient.ClusterRoleBindingApply(kubernetesCtx, log, o)

		case *rbacv1.Role:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "Role", "name", o.Name)
			errRes = client.k8sClient.RoleApply(kubernetesCtx, log, o)

		case *rbacv1.RoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "RoleBinding", "name", o.Name)
			errRes = client.k8sClient.RoleBindingApply(kubernetesCtx, log, o)

		case *networkingv1.NetworkPolicy:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			log.Print(kubernetesCtx, "NetworkPolicy", "name", o.Name)
			errRes = client.k8sClient.NetPolicyApply(kubernetesCtx, log, o)

		default:
			errRes = ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "unexpected type", "obj", o),
			)
		}

		if errRes != nil {
			return errRes
		}
	}

	return nil
}
