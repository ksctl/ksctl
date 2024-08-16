package kubernetes

import (
	"bytes"
	"fmt"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"gopkg.in/yaml.v3"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"net/http"
	"os"
	"strings"
)

func (k *K8sClusterClient) getDynamicClientFromManifest(manifest []byte) (dynamic.ResourceInterface, *unstructured.Unstructured, error) {

	dynamicClient, err := dynamic.NewForConfig(k.c)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating dynamic client: %w", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(k.c)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %v", err)
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get API group resources: %v", err)
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	obj := new(unstructured.Unstructured)
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	if err := decoder.Decode(&obj); err != nil {
		return nil, nil, fmt.Errorf("error decoding manifest: %w", err)
	}

	gvk := obj.GroupVersionKind()

	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get REST mapping: %v", err)
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default" // default namespace if not specified
	}

	var dri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dri = dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		dri = dynamicClient.Resource(mapping.Resource)
	}

	return dri, obj, nil
}

func httpGet(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to get manifests", "Reason", err),
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to get manifests", "Got StatusCode", resp.StatusCode),
		)
	}

	r_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to read the manifests", "Reason", err),
		)
	}

	return r_body, nil
}

func fileGet(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to open the manifest location", "Reason", err),
		)
	}

	defer f.Close()

	r_body, err := io.ReadAll(f)
	if err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to read the manifests", "Reason", err),
		)
	}

	return r_body, nil
}

func Http(uri string) (string, error) {
	var v []byte
	var err error
	if strings.HasPrefix(uri, "uri:::") {
		v, err = httpGet(strings.TrimPrefix(uri, "uri:::"))
	} else if strings.HasPrefix(uri, "file:::") {
		v, err = fileGet(strings.TrimPrefix(uri, "file:::"))
	} else {
		v, err = httpGet(uri)
	}
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func splitYAMLDocuments(content string) ([]string, error) {
	decoder := yaml.NewDecoder(strings.NewReader(content))
	var documents []string

	for {
		var doc interface{}
		err := decoder.Decode(&doc)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		docBytes, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		documents = append(documents, string(docBytes))
	}

	return documents, nil
}

func getManifests(appUrl string) ([]string, error) {
	body, err := Http(appUrl)
	if err != nil {
		return nil, err
	}

	if err := apiextensionsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "failed to add apiextensionv1 to the scheme", "Reason", err),
		)
	}
	return splitYAMLDocuments(body)
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
			log.Warn(kubernetesCtx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err)

			if c, o, err := client.getDynamicClientFromManifest([]byte(resource)); err != nil {
				return err
			} else {
				err := c.Delete(kubernetesCtx, o.GetName(), metav1.DeleteOptions{})
				if err != nil {
					return err
				}
				continue
			}
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

		case *nodev1.RuntimeClass:
			log.Print(kubernetesCtx, "RuntimeClass", "name", o.Name)
			errRes = client.k8sClient.RuntimeDelete(kubernetesCtx, log, o.Name)

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
			ObjectMeta: metav1.ObjectMeta{
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
			ObjectMeta: metav1.ObjectMeta{
				Name: component.Namespace,
			}}); err != nil {
			return err
		}
	}

	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			log.Warn(kubernetesCtx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err)

			if c, o, err := client.getDynamicClientFromManifest([]byte(resource)); err != nil {
				return err
			} else {
				_, err := c.Create(kubernetesCtx, o, metav1.CreateOptions{})
				if err != nil {
					return err
				}
				continue
			}
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

		case *nodev1.RuntimeClass:
			log.Print(kubernetesCtx, "RuntimeClass", "name", o.Name)
			errRes = client.k8sClient.RuntimeApply(kubernetesCtx, log, o)

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
