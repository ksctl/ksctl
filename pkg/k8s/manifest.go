package k8s

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"

	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
)

func (k *Client) getDynamicClientFromManifest(manifest []byte) (dynamic.ResourceInterface, *unstructured.Unstructured, error) {

	dynamicClient, err := dynamic.NewForConfig(k.r)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating dynamic client: %w", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(k.r)
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

func (k *Client) httpGet(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to get manifests", "Reason", err),
		)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to get manifests", "Got StatusCode", resp.StatusCode),
		)
	}

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to read the manifests", "Reason", err),
		)
	}

	return rBody, nil
}

func (k *Client) fileGet(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to open the manifest location", "Reason", err),
		)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	rBody, err := io.ReadAll(f)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to read the manifests", "Reason", err),
		)
	}

	return rBody, nil
}

func (k *Client) getManifest(uri string) (string, error) {
	var v []byte
	var err error
	if strings.HasPrefix(uri, "uri:::") {
		v, err = k.httpGet(strings.TrimPrefix(uri, "uri:::"))
	} else if strings.HasPrefix(uri, "file:::") {
		v, err = k.fileGet(strings.TrimPrefix(uri, "file:::"))
	} else {
		v, err = k.httpGet(uri)
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

func (k *Client) getManifests(appUrl string) ([]string, error) {
	body, err := k.getManifest(appUrl)
	if err != nil {
		return nil, err
	}

	if err := apiextensionsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to add apiextensionv1 to the scheme", "Reason", err),
		)
	}
	return splitYAMLDocuments(body)
}

func (k *Client) KubectlApply(component *App) error {

	for _, url := range component.Urls {
		resources, err := k.getManifests(url)
		if err != nil {
			return err
		}
		if err := k.individualKubectlInstall(component, resources); err != nil {
			return err
		}
	}
	return nil
}

func (k *Client) individualKubectlInstall(component *App, resources []string) error {
	if component.CreateNamespace {
		if err := k.NamespaceCreate(&corev1.Namespace{
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
			k.l.Warn(k.ctx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err)

			if c, o, err := k.getDynamicClientFromManifest([]byte(resource)); err != nil {
				return err
			} else {
				_, err := c.Create(k.ctx, o, metav1.CreateOptions{})
				if err != nil {
					return err
				}
				continue
			}
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = k.ApiExtensionsApply(o)

		case *corev1.Namespace:
			k.l.Print(k.ctx, "Namespace", "name", o.Name)
			errRes = k.NamespaceCreate(o)

		case *appsv1.DaemonSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Daemonset", "name", o.Name)
			errRes = k.DaemonsetApply(o)

		case *appsv1.Deployment:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Deployment", "name", o.Name)
			errRes = k.DeploymentApply(o)

		case *corev1.Service:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Service", "name", o.Name)
			errRes = k.ServiceApply(o)

		case *corev1.ServiceAccount:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ServiceAccount", "name", o.Name)
			errRes = k.ServiceAccountApply(o)

		case *corev1.ConfigMap:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ConfigMap", "name", o.Name)
			errRes = k.ConfigMapApply(o)

		case *corev1.Secret:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Secret", "name", o.Name)
			errRes = k.SecretApply(o)

		case *nodev1.RuntimeClass:
			k.l.Print(k.ctx, "RuntimeClass", "name", o.Name)
			errRes = k.RuntimeApply(o)

		case *appsv1.StatefulSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "StatefulSet", "name", o.Name)
			errRes = k.StatefulSetApply(o)

		case *rbacv1.ClusterRole:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ClusterRole", "name", o.Name)
			errRes = k.ClusterRoleApply(o)

		case *rbacv1.ClusterRoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ClusterRoleBinding", "name", o.Name)
			errRes = k.ClusterRoleBindingApply(o)

		case *rbacv1.Role:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Role", "name", o.Name)
			errRes = k.RoleApply(o)

		case *rbacv1.RoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "RoleBinding", "name", o.Name)
			errRes = k.RoleBindingApply(o)

		case *networkingv1.NetworkPolicy:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "NetworkPolicy", "name", o.Name)
			errRes = k.NetPolicyApply(o)

		default:
			errRes = ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "unexpected type", "obj", o),
			)
		}

		if errRes != nil {
			return errRes
		}
	}

	return nil
}

func (k *Client) KubectlDelete(component *App) error {
	for idx := len(component.Urls) - 1; idx >= 0; idx-- {
		resources, err := k.getManifests(component.Urls[idx])
		if err != nil {
			return err
		}
		if err := k.individualKubectlUninstall(component, resources); err != nil {
			return err
		}
	}
	return nil
}

func (k *Client) individualKubectlUninstall(component *App, resources []string) error {
	for _, resource := range resources {
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			k.l.Warn(k.ctx, "failed to decode the raw manifests into kubernetes gvr", "Reason", err)

			if c, o, err := k.getDynamicClientFromManifest([]byte(resource)); err != nil {
				return err
			} else {
				err := c.Delete(k.ctx, o.GetName(), metav1.DeleteOptions{})
				if err != nil {
					return err
				}
				continue
			}
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = k.ApiExtensionsDelete(o)

		case *corev1.Namespace:
			k.l.Print(k.ctx, "Namespace", "name", o.Name)
			errRes = k.NamespaceDelete(o, false)

		case *appsv1.DaemonSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Daemonset", "name", o.Name)
			errRes = k.DaemonsetDelete(o)

		case *appsv1.Deployment:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Deployment", "name", o.Name)
			errRes = k.DeploymentDelete(o)

		case *corev1.Service:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Service", "name", o.Name)
			errRes = k.ServiceDelete(o)

		case *corev1.ServiceAccount:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ServiceAccount", "name", o.Name)
			errRes = k.ServiceAccountDelete(o)

		case *corev1.ConfigMap:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ConfigMap", "name", o.Name)
			errRes = k.ConfigMapDelete(o)

		case *corev1.Secret:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Secret", "name", o.Name)
			errRes = k.SecretDelete(o)

		case *appsv1.StatefulSet:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "StatefulSet", "name", o.Name)
			errRes = k.StatefulSetDelete(o)

		case *nodev1.RuntimeClass:
			k.l.Print(k.ctx, "RuntimeClass", "name", o.Name)
			errRes = k.RuntimeDelete(o.Name)

		case *rbacv1.ClusterRole:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ClusterRole", "name", o.Name)
			errRes = k.ClusterRoleDelete(o)

		case *rbacv1.ClusterRoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "ClusterRoleBinding", "name", o.Name)
			errRes = k.ClusterRoleBindingDelete(o)

		case *rbacv1.Role:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "Role", "name", o.Name)
			errRes = k.RoleDelete(o)

		case *rbacv1.RoleBinding:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "RoleBinding", "name", o.Name)
			errRes = k.RoleBindingDelete(o)

		case *networkingv1.NetworkPolicy:
			if component.CreateNamespace {
				o.Namespace = component.Namespace
			}
			k.l.Print(k.ctx, "NetworkPolicy", "name", o.Name)
			errRes = k.NetPolicyDelete(o)

		default:
			errRes = ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "unexpected type", "obj", o),
			)
		}

		if errRes != nil {
			return errRes
		}
	}

	if component.CreateNamespace {
		if err := k.NamespaceDelete(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: component.Namespace,
			}}, true); err != nil {
			return err
		}
	}

	return nil
}
