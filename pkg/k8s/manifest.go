// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"gopkg.in/yaml.v3"
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

func (k *Client) getDynamicClientFromManifest(manifest []byte, defaultNamespace *string) (dynamic.ResourceInterface, *unstructured.Unstructured, error) {

	dynamicClient, err := dynamic.NewForConfig(k.R)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating dynamic client: %w", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(k.R)
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
		if defaultNamespace != nil {
			namespace = *defaultNamespace
		} else {
			namespace = "default"
		}
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

func (k *Client) getManifest(uri string) (string, error) {
	var v []byte
	var err error
	v, err = k.httpGet(uri)
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
		ns := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": component.Namespace,
				},
			},
		}

		if c, _, err := k.getDynamicClientFromManifest([]byte(mustYAMLMarshal(ns)), nil); err != nil {
			return err
		} else {
			if _, err := c.Apply(k.ctx, ns.GetName(), ns, metav1.ApplyOptions{
				FieldManager: "ksctl",
				Force:        true,
			}); err != nil {
				return err
			}
			k.l.Success(k.ctx, "Created Namespace", "name", component.Namespace)
		}
	}

	for _, resource := range resources {
		ns := func() *string {
			if component.CreateNamespace {
				return &component.Namespace
			}
			return nil
		}()

		c, obj, err := k.getDynamicClientFromManifest([]byte(resource), ns)
		if err != nil {
			return err
		}

		k.l.Print(k.ctx, "Applying Resource", "kind", obj.GetKind(), "name", obj.GetName())

		_, err = c.Apply(k.ctx, obj.GetName(), obj, metav1.ApplyOptions{
			FieldManager: "ksctl",
			Force:        true,
		})
		if err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}

		k.l.Success(k.ctx, "Applied Resource", "kind", obj.GetKind(), "name", obj.GetName(), "namespace", obj.GetNamespace())
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
		ns := func() *string {
			if component.CreateNamespace {
				return &component.Namespace
			}
			return nil
		}()

		c, obj, err := k.getDynamicClientFromManifest([]byte(resource), ns)
		if err != nil {
			return err
		}

		k.l.Print(k.ctx, "Deleting Resource", "kind", obj.GetKind(), "name", obj.GetName())

		err = c.Delete(k.ctx, obj.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete resource %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}

		k.l.Success(k.ctx, "Deleted Resource", "kind", obj.GetKind(), "name", obj.GetName(), "namespace", obj.GetNamespace())
	}

	if component.CreateNamespace {
		// Delete namespace using dynamic client
		ns := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": component.Namespace,
				},
			},
		}

		if c, _, err := k.getDynamicClientFromManifest([]byte(mustYAMLMarshal(ns)), nil); err != nil {
			return err
		} else {
			if err := c.Delete(k.ctx, ns.GetName(), metav1.DeleteOptions{}); err != nil {
				return err
			}
			k.l.Success(k.ctx, "Deleted Namespace", "name", component.Namespace)
		}
	}

	return nil
}

// Helper function to marshal unstructured objects to YAML
func mustYAMLMarshal(obj interface{}) []byte {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return bytes
}
