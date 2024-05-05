package kubernetes

import (
	"io"
	"net/http"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type KubectlOptions struct {
	createNamespace bool
}

func getManifests(app Application) ([]string, error) {

	// Get the manifest
	resp, err := http.Get(app.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Split the manifest into individual resources
	resources := strings.Split(string(body), "---")
	if err := apiextensionsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	return resources, nil
}

func deleteKubectl(client *Kubernetes, appStruct Application) error {
	resources, err := getManifests(appStruct)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			return err
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = client.apiExtensionsDelete(o)

		case *corev1.Namespace:
			log.Debug("Namespace", "name", o.Name)
			errRes = client.namespaceDeleteManifest(o)

		case *appsv1.Deployment:

			log.Debug("Deployment", "name", o.Name)
			errRes = client.deploymentDelete(o, appStruct.Namespace)

		case *corev1.Service:
			log.Debug("Service", "name", o.Name)
			errRes = client.serviceDelete(o, appStruct.Namespace)

		case *corev1.ServiceAccount:
			log.Debug("ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountDelete(o, appStruct.Namespace)

		case *corev1.ConfigMap:
			log.Debug("ConfigMap", "name", o.Name)
			errRes = client.configMapDelete(o, appStruct.Namespace)

		case *corev1.Secret:
			log.Debug("Secret", "name", o.Name)
			errRes = client.secretDelete(o, appStruct.Namespace)

		case *appsv1.StatefulSet:
			log.Debug("StatefulSet", "name", o.Name)
			errRes = client.statefulSetDelete(o, appStruct.Namespace)

		case *rbacv1.ClusterRole:
			log.Debug("ClusterRole", "name", o.Name)
			errRes = client.clusterRoleDelete(o)

		case *rbacv1.ClusterRoleBinding:
			log.Debug("ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingDelete(o)

		case *rbacv1.Role:
			log.Debug("Role", "name", o.Name)
			errRes = client.roleDelete(o, appStruct.Namespace)

		case *rbacv1.RoleBinding:
			log.Debug("RoleBinding", "name", o.Name)
			errRes = client.roleBindingDelete(o, appStruct.Namespace)

		case *networkingv1.NetworkPolicy:
			log.Debug("NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyDelete(o, appStruct.Namespace)

		default:
			log.Error("unexpected type", "obj", o)
		}

		if errRes != nil {
			return errRes
		}
	}

	if appStruct.KubectlConfig.createNamespace {
		if err := client.namespaceDelete(appStruct.Namespace); err != nil {
			return err
		}
	}

	return nil
}

func installKubectl(client *Kubernetes, appStruct Application) error {
	resources, err := getManifests(appStruct)
	if err != nil {
		return err
	}

	if appStruct.KubectlConfig.createNamespace {
		if err := client.namespaceCreate(appStruct.Namespace); err != nil {
			return err
		}
	}

	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			return err
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = client.apiExtensionsApply(o)

		case *corev1.Namespace:
			log.Debug("Namespace", "name", o.Name)
			errRes = client.namespaceCreateManifest(o)

		case *appsv1.Deployment:

			log.Debug("Deployment", "name", o.Name)
			errRes = client.deploymentApply(o, appStruct.Namespace)

		case *corev1.Service:
			log.Debug("Service", "name", o.Name)
			errRes = client.serviceApply(o, appStruct.Namespace)

		case *corev1.ServiceAccount:
			log.Debug("ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountApply(o, appStruct.Namespace)

		case *corev1.ConfigMap:
			log.Debug("ConfigMap", "name", o.Name)
			errRes = client.configMapApply(o, appStruct.Namespace)

		case *corev1.Secret:
			log.Debug("Secret", "name", o.Name)
			errRes = client.secretApply(o, appStruct.Namespace)

		case *appsv1.StatefulSet:
			log.Debug("StatefulSet", "name", o.Name)
			errRes = client.statefulSetApply(o, appStruct.Namespace)

		case *rbacv1.ClusterRole:
			log.Debug("ClusterRole", "name", o.Name)
			errRes = client.clusterRoleApply(o)

		case *rbacv1.ClusterRoleBinding:
			log.Debug("ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingApply(o)

		case *rbacv1.Role:
			log.Debug("Role", "name", o.Name)
			errRes = client.roleApply(o, appStruct.Namespace)

		case *rbacv1.RoleBinding:
			log.Debug("RoleBinding", "name", o.Name)
			errRes = client.roleBindingApply(o, appStruct.Namespace)

		case *networkingv1.NetworkPolicy:
			log.Debug("NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyApply(o, appStruct.Namespace)

		default:
			log.Error("unexpected type", "obj", o)
		}

		if errRes != nil {
			return errRes
		}
	}

	return nil
}
