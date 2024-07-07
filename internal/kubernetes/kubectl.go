package kubernetes

import (
	"io"
	"net/http"
	"strings"

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

func deleteKubectl(client *Kubernetes, component *KubectlHandler) error {
	resources, err := getManifests(component.url)
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
			errRes = client.apiExtensionsDelete(o)

		case *corev1.Namespace:
			log.Print(kubernetesCtx, "Namespace", "name", o.Name)
			errRes = client.namespaceDelete(o, false)

		case *appsv1.DaemonSet:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Daemonset", "name", o.Name)
			errRes = client.daemonsetDelete(o)

		case *appsv1.Deployment:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Deployment", "name", o.Name)
			errRes = client.deploymentDelete(o)

		case *corev1.Service:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Service", "name", o.Name)
			errRes = client.serviceDelete(o)

		case *corev1.ServiceAccount:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountDelete(o)

		case *corev1.ConfigMap:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ConfigMap", "name", o.Name)
			errRes = client.configMapDelete(o)

		case *corev1.Secret:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Secret", "name", o.Name)
			errRes = client.secretDelete(o)

		case *appsv1.StatefulSet:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "StatefulSet", "name", o.Name)
			errRes = client.statefulSetDelete(o)

		case *rbacv1.ClusterRole:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ClusterRole", "name", o.Name)
			errRes = client.clusterRoleDelete(o)

		case *rbacv1.ClusterRoleBinding:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingDelete(o)

		case *rbacv1.Role:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Role", "name", o.Name)
			errRes = client.roleDelete(o)

		case *rbacv1.RoleBinding:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "RoleBinding", "name", o.Name)
			errRes = client.roleBindingDelete(o)

		case *networkingv1.NetworkPolicy:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyDelete(o)

		default:
			errRes = ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "unexpected type", "obj", o),
			)
		}

		if errRes != nil {
			return errRes
		}
	}

	if component.createNamespace {
		if err := client.namespaceDelete(&corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: component.namespace,
			}}, true); err != nil {
			return err
		}
	}

	return nil
}

func installKubectl(client *Kubernetes, component *KubectlHandler) error {
	resources, err := getManifests(component.url)
	if err != nil {
		return err
	}

	if component.createNamespace {
		if err := client.namespaceCreate(&corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: component.namespace,
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
			errRes = client.apiExtensionsApply(o)

		case *corev1.Namespace:
			log.Print(kubernetesCtx, "Namespace", "name", o.Name)
			errRes = client.namespaceCreate(o)

		case *appsv1.DaemonSet:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Daemonset", "name", o.Name)
			errRes = client.daemonsetApply(o)

		case *appsv1.Deployment:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Deployment", "name", o.Name)
			errRes = client.deploymentApply(o)

		case *corev1.Service:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Service", "name", o.Name)
			errRes = client.serviceApply(o)

		case *corev1.ServiceAccount:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountApply(o)

		case *corev1.ConfigMap:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ConfigMap", "name", o.Name)
			errRes = client.configMapApply(o)

		case *corev1.Secret:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Secret", "name", o.Name)
			errRes = client.secretApply(o)

		case *appsv1.StatefulSet:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "StatefulSet", "name", o.Name)
			errRes = client.statefulSetApply(o)

		case *rbacv1.ClusterRole:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ClusterRole", "name", o.Name)
			errRes = client.clusterRoleApply(o)

		case *rbacv1.ClusterRoleBinding:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingApply(o)

		case *rbacv1.Role:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "Role", "name", o.Name)
			errRes = client.roleApply(o)

		case *rbacv1.RoleBinding:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "RoleBinding", "name", o.Name)
			errRes = client.roleBindingApply(o)

		case *networkingv1.NetworkPolicy:
			if component.createNamespace {
				o.Namespace = component.namespace
			}
			log.Print(kubernetesCtx, "NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyApply(o)

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
