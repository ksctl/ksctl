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

package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type ClientSet interface {
	WriteConfigMap(namespace string, c *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
	WriteSecret(namespace string, c *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error)

	ReadSecret(namespace, name string, opts metav1.GetOptions) (*v1.Secret, error)
	ReadConfigMap(namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error)
}

type Client struct {
	client *kubernetes.Clientset
	ctx    context.Context
}

func NewK8sClient(ctx context.Context) (ClientSet, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "Error loading in-cluster config", "Reason", err),
		)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "Error creating Kubernetes client", "Reason", err),
		)
	}

	return &Client{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c2 *Client) WriteConfigMap(namespace string, c *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	if ret, err := c2.client.CoreV1().ConfigMaps(namespace).Update(c2.ctx, c, opts); err != nil {
		if errors.IsNotFound(err) {
			return c2.client.CoreV1().ConfigMaps(namespace).Create(c2.ctx, c, metav1.CreateOptions{})
		} else {
			return nil, err
		}

	} else {
		return ret, err
	}
}

func (c2 *Client) WriteSecret(namespace string, s *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	if ret, err := c2.client.CoreV1().Secrets(namespace).Update(c2.ctx, s, opts); err != nil {
		if errors.IsNotFound(err) {
			return c2.client.CoreV1().Secrets(namespace).Create(c2.ctx, s, metav1.CreateOptions{})
		} else {
			return nil, err
		}

	} else {
		return ret, err
	}
}

func (c2 *Client) ReadSecret(namespace, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	return c2.client.CoreV1().Secrets(namespace).Get(c2.ctx, name, opts)
}

func (c2 *Client) ReadConfigMap(namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	return c2.client.CoreV1().ConfigMaps(namespace).Get(c2.ctx, name, opts)
}

type FakeClient struct {
	client *fake.Clientset
	ctx    context.Context
}

func NewFakeK8sClient(ctx context.Context) (ClientSet, error) {
	return &FakeClient{
		ctx:    ctx,
		client: fake.NewSimpleClientset(),
	}, nil
}

func (f *FakeClient) WriteConfigMap(namespace string, c *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	if ret, err := f.client.CoreV1().ConfigMaps(namespace).Update(f.ctx, c, opts); err != nil {
		if errors.IsNotFound(err) {
			return f.client.CoreV1().ConfigMaps(namespace).Create(f.ctx, c, metav1.CreateOptions{})
		} else {
			return nil, err
		}

	} else {
		return ret, err
	}
}

func (f *FakeClient) WriteSecret(namespace string, s *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	if ret, err := f.client.CoreV1().Secrets(namespace).Update(f.ctx, s, opts); err != nil {
		if errors.IsNotFound(err) {
			return f.client.CoreV1().Secrets(namespace).Create(f.ctx, s, metav1.CreateOptions{})
		} else {
			return nil, err
		}

	} else {
		return ret, err
	}
}

func (f *FakeClient) ReadSecret(namespace, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	return f.client.CoreV1().Secrets(namespace).Get(f.ctx, name, opts)
}

func (f *FakeClient) ReadConfigMap(namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	return f.client.CoreV1().ConfigMaps(namespace).Get(f.ctx, name, opts)
}

func (f *FakeClient) DeleteResourcesForTest() error {
	if err := f.client.
		CoreV1().
		ConfigMaps(ksctlNamespace).
		Delete(f.ctx, ksctlStateName, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
