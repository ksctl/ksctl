package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type KubectlOptions struct {
	createNamespace bool

	// Namespace Only specify if createNamespace is true
	namespace string
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
			log.Print("Namespace", "name", o.Name)
			errRes = client.namespaceDelete(o, false)

		case *appsv1.DaemonSet:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Daemonset", "name", o.Name)
			errRes = client.daemonsetDelete(o)

		case *appsv1.Deployment:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Deployment", "name", o.Name)
			errRes = client.deploymentDelete(o)

		case *corev1.Service:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Service", "name", o.Name)
			errRes = client.serviceDelete(o)

		case *corev1.ServiceAccount:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountDelete(o)

		case *corev1.ConfigMap:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ConfigMap", "name", o.Name)
			errRes = client.configMapDelete(o)

		case *corev1.Secret:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Secret", "name", o.Name)
			errRes = client.secretDelete(o)

		case *appsv1.StatefulSet:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("StatefulSet", "name", o.Name)
			errRes = client.statefulSetDelete(o)

		case *rbacv1.ClusterRole:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ClusterRole", "name", o.Name)
			errRes = client.clusterRoleDelete(o)

		case *rbacv1.ClusterRoleBinding:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingDelete(o)

		case *rbacv1.Role:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Role", "name", o.Name)
			errRes = client.roleDelete(o)

		case *rbacv1.RoleBinding:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("RoleBinding", "name", o.Name)
			errRes = client.roleBindingDelete(o)

		case *networkingv1.NetworkPolicy:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyDelete(o)

		default:
			log.Error("unexpected type", "obj", o)
		}

		if errRes != nil {
			return errRes
		}
	}

	if appStruct.KubectlConfig.createNamespace {
		if err := client.namespaceDelete(&corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: appStruct.KubectlConfig.namespace,
			}}, true); err != nil {
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
		if err := client.namespaceCreate(&corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: appStruct.KubectlConfig.namespace,
			}}); err != nil {
			return err
		}
	}

	for _, resource := range resources {
		fmt.Println(resource)
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
			log.Print("Namespace", "name", o.Name)
			errRes = client.namespaceCreate(o)

		case *appsv1.DaemonSet:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Daemonset", "name", o.Name)
			errRes = client.daemonsetApply(o)

		case *appsv1.Deployment:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Deployment", "name", o.Name)
			errRes = client.deploymentApply(o)

		case *corev1.Service:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Service", "name", o.Name)
			errRes = client.serviceApply(o)

		case *corev1.ServiceAccount:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ServiceAccount", "name", o.Name)
			errRes = client.serviceAccountApply(o)

		case *corev1.ConfigMap:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ConfigMap", "name", o.Name)
			errRes = client.configMapApply(o)

		case *corev1.Secret:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Secret", "name", o.Name)
			errRes = client.secretApply(o)

		case *appsv1.StatefulSet:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("StatefulSet", "name", o.Name)
			errRes = client.statefulSetApply(o)

		case *rbacv1.ClusterRole:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ClusterRole", "name", o.Name)
			errRes = client.clusterRoleApply(o)

		case *rbacv1.ClusterRoleBinding:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("ClusterRoleBinding", "name", o.Name)
			errRes = client.clusterRoleBindingApply(o)

		case *rbacv1.Role:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("Role", "name", o.Name)
			errRes = client.roleApply(o)

		case *rbacv1.RoleBinding:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("RoleBinding", "name", o.Name)
			errRes = client.roleBindingApply(o)

		case *networkingv1.NetworkPolicy:
			if appStruct.KubectlConfig.createNamespace {
				o.Namespace = appStruct.KubectlConfig.namespace
			}
			log.Print("NetworkPolicy", "name", o.Name)
			errRes = client.netPolicyApply(o)

		default:
			log.Error("unexpected type", "obj", o)
		}

		if errRes != nil {
			return errRes
		}
	}

	return nil
}
