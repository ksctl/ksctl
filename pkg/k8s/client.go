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
	"context"

	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	ctx                 context.Context
	l                   logger.Logger
	clientset           kubernetes.Interface
	apiextensionsClient clientset.Interface
	r                   *rest.Config
}

type App struct {
	CreateNamespace bool
	Namespace       string

	Version     string   // how to get the version propogated?
	Urls        []string // make sure you traverse backwards to uninstall resources
	PostInstall string
	Metadata    string
}

func NewK8sClient(ctx context.Context, l logger.Logger, c *rest.Config) (k *Client, err error) {
	k = new(Client)
	k.ctx = context.WithValue(ctx, "module", "kubernetes-client")
	k.r = c
	k.l = l
	k.apiextensionsClient, err = clientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	k.clientset, err = kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// // ClusterRoleApply creates or updates a ClusterRole
// func (k *Client) ClusterRoleApply(cr *rbacv1.ClusterRole) error {
// 	k.l.Debug(k.ctx, "Creating/updating ClusterRole", "name", cr.Name)

// 	_, err := k.clientset.RbacV1().ClusterRoles().Get(context.Background(), cr.Name, metav1.GetOptions{})
// 	if err != nil {
// 		if errors.IsNotFound(err) {
// 			// ClusterRole doesn't exist, create it
// 			_, err = k.clientset.RbacV1().ClusterRoles().Create(context.Background(), cr, metav1.CreateOptions{})
// 			if err != nil {
// 				return err
// 			}
// 			k.l.Debug(k.ctx, "Created ClusterRole", "name", cr.Name)
// 			return nil
// 		}
// 		return err
// 	}

// 	// ClusterRole exists, update it
// 	_, err = k.clientset.RbacV1().ClusterRoles().Update(context.Background(), cr, metav1.UpdateOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	k.l.Debug(k.ctx, "Updated ClusterRole", "name", cr.Name)
// 	return nil
// }

// // ClusterRoleBindingApply creates or updates a ClusterRoleBinding
// func (k *Client) ClusterRoleBindingApply(crb *rbacv1.ClusterRoleBinding) error {
// 	k.l.Debug(k.ctx, "Creating/updating ClusterRoleBinding", "name", crb.Name)

// 	_, err := k.clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), crb.Name, metav1.GetOptions{})
// 	if err != nil {
// 		if errors.IsNotFound(err) {
// 			// ClusterRoleBinding doesn't exist, create it
// 			_, err = k.clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), crb, metav1.CreateOptions{})
// 			if err != nil {
// 				return err
// 			}
// 			k.l.Debug(k.ctx, "Created ClusterRoleBinding", "name", crb.Name)
// 			return nil
// 		}
// 		return err
// 	}

// 	// ClusterRoleBinding exists, update it
// 	_, err = k.clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), crb, metav1.UpdateOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	k.l.Debug(k.ctx, "Updated ClusterRoleBinding", "name", crb.Name)
// 	return nil
// }

// CreateServiceAccountToken creates a token for a ServiceAccount
func (k *Client) CreateServiceAccountToken(name string, namespace string) (*authenticationv1.TokenRequest, error) {
	k.l.Debug(k.ctx, "Creating token for ServiceAccount", "name", name, "namespace", namespace)

	// Let's make sure the service account exists first
	_, err := k.clientset.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Create a token that never expires (or set expirationSeconds to desired TTL)
	tokenRequest := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: utilities.Ptr(int64(86400)), // 24 hours by default
		},
	}

	tokenResponse, err := k.clientset.CoreV1().ServiceAccounts(namespace).CreateToken(
		context.Background(),
		name,
		tokenRequest,
		metav1.CreateOptions{},
	)

	if err != nil {
		return nil, err
	}

	k.l.Debug(k.ctx, "Created token for ServiceAccount", "name", name, "namespace", namespace)
	return tokenResponse, nil
}
