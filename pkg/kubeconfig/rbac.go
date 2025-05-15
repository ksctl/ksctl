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

package kubeconfig

import (
	"fmt"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Role names
	adminRoleName      = "ksctl-admin"
	maintainerRoleName = "ksctl-maintainer"
	developerRoleName  = "ksctl-developer"

	// ServiceAccount names
	adminSAName      = "ksctl-admin"
	maintainerSAName = "ksctl-maintainer"
	developerSAName  = "ksctl-developer"

	// Namespace
	ksctlNamespace = "ksctl-rbac"
)

func (g *Generator) configureAdminRole(k8sClient *k8s.Client) error {
	g.log.Debug(g.ctx, "Configuring admin role")

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      adminSAName,
			Namespace: ksctlNamespace,
		},
	}

	// Create namespace if it doesn't exist
	if err := k8sClient.NamespaceCreate(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ksctlNamespace,
		},
	}); err != nil {
		g.log.Debug(g.ctx, "Namespace already exists or error creating", "error", err)
	}

	// Create ServiceAccount
	if err := k8sClient.ServiceAccountApply(sa); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create admin service account: %v", err,
		)
	}

	// Create ClusterRoleBinding to cluster-admin
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ksctl-admin-binding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin", // Use existing cluster-admin role
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      adminSAName,
				Namespace: ksctlNamespace,
			},
		},
	}

	if err := k8sClient.ClusterRoleBindingApply(crb); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create admin cluster role binding: %v", err,
		)
	}

	g.log.Success(g.ctx, "Configured admin role")
	return nil
}

func (g *Generator) configureMaintainerRole(k8sClient *k8s.Client) error {
	g.log.Debug(g.ctx, "Configuring maintainer role")

	// Create namespace if it doesn't exist
	if err := k8sClient.NamespaceCreate(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ksctlNamespace,
		},
	}); err != nil {
		g.log.Debug(g.ctx, "Namespace already exists or error creating", "error", err)
	}

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      maintainerSAName,
			Namespace: ksctlNamespace,
		},
	}

	if err := k8sClient.ServiceAccountApply(sa); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create maintainer service account: %v", err,
		)
	}

	// Create ClusterRole for maintainer
	// A maintainer can manage workloads but not modify cluster infrastructure
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: maintainerRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "configmaps", "secrets", "persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "statefulsets", "daemonsets", "replicasets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses", "networkpolicies"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list", "watch", "create"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"roles", "rolebindings"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			// Read-only access to nodes and cluster state
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "persistentvolumes", "events"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	if err := k8sClient.ClusterRoleApply(cr); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create maintainer cluster role: %v", err,
		)
	}

	// Create ClusterRoleBinding for maintainer
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ksctl-maintainer-binding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     maintainerRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      maintainerSAName,
				Namespace: ksctlNamespace,
			},
		},
	}

	if err := k8sClient.ClusterRoleBindingApply(crb); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create maintainer cluster role binding: %v", err,
		)
	}

	g.log.Success(g.ctx, "Configured maintainer role")
	return nil
}

func (g *Generator) configureDeveloperRole(k8sClient *k8s.Client) error {
	g.log.Debug(g.ctx, "Configuring developer role")

	// Create namespace if it doesn't exist
	if err := k8sClient.NamespaceCreate(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ksctlNamespace,
		},
	}); err != nil {
		g.log.Debug(g.ctx, "Namespace already exists or error creating", "error", err)
	}

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      developerSAName,
			Namespace: ksctlNamespace,
		},
	}

	if err := k8sClient.ServiceAccountApply(sa); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create developer service account: %v", err,
		)
	}

	// Create ClusterRole for developer
	// A developer has limited access, focused on their namespaces
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: developerRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "configmaps", "secrets", "persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "statefulsets", "daemonsets", "replicasets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses", "networkpolicies"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			// Read-only access to namespaces
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
			// Read-only access to cluster state
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "persistentvolumes", "events"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	if err := k8sClient.ClusterRoleApply(cr); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create developer cluster role: %v", err,
		)
	}

	// Create ClusterRoleBinding for developer
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ksctl-developer-binding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     developerRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      developerSAName,
				Namespace: ksctlNamespace,
			},
		},
	}

	if err := k8sClient.ClusterRoleBindingApply(crb); err != nil {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create developer cluster role binding: %v", err,
		)
	}

	g.log.Success(g.ctx, "Configured developer role")
	return nil
}

// GetServiceAccountData retrieves the service account token for a specific role
func (g *Generator) GetServiceAccountData(k8sClient *k8s.Client, role consts.KsctlKubeRole) (string, error) {
	var saName string

	switch role {
	case consts.KubeRoleAdmin:
		saName = adminSAName
	case consts.KubeRoleMaintainer:
		saName = maintainerSAName
	case consts.KubeRoleDeveloper:
		saName = developerSAName
	default:
		return "", fmt.Errorf("unsupported role: %s", role)
	}

	// Create the token request
	tokenRequest, err := k8sClient.CreateServiceAccountToken(saName, ksctlNamespace)
	if err != nil {
		return "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"failed to create token for service account %s: %v", saName, err,
		)
	}

	return tokenRequest.Status.Token, nil
}
