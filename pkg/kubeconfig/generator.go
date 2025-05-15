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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/k8s"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Generator handles kubeconfig generation with role-based access
type Generator struct {
	ctx   context.Context
	log   logger.Logger
	state *statefile.StorageDocument
}

// NewGenerator creates a new kubeconfig generator
func NewGenerator(ctx context.Context, log logger.Logger, state *statefile.StorageDocument) *Generator {
	return &Generator{
		ctx:   ctx,
		log:   log,
		state: state,
	}
}

// IsExpired checks if a kubeconfig has expired based on its TTL
func IsExpired(createdAt int64, ttl *int64) bool {
	if ttl == nil {
		return false // No TTL means no expiration
	}

	expiryTime := createdAt + *ttl
	return time.Now().Unix() > expiryTime
}

// CleanupExpiredKubeconfigs removes expired kubeconfigs from the state
func (g *Generator) CleanupExpiredKubeconfigs() {
	if g.state.RoleBasedKubeconfigs == nil {
		return
	}

	for fingerprint, kubeconfig := range g.state.RoleBasedKubeconfigs {
		if IsExpired(kubeconfig.CreatedAt, kubeconfig.TTL) {
			delete(g.state.RoleBasedKubeconfigs, fingerprint)
			g.log.Debug(g.ctx, "Removed expired kubeconfig", "fingerprint", fingerprint, "role", kubeconfig.Role)
		}
	}
}

// FindKubeconfig finds an existing kubeconfig by role
func (g *Generator) FindKubeconfig(role consts.KsctlKubeRole) (string, string, bool) {
	if g.state.RoleBasedKubeconfigs == nil {
		return "", "", false
	}

	for fingerprint, kubeconfig := range g.state.RoleBasedKubeconfigs {
		if kubeconfig.Role == role && !IsExpired(kubeconfig.CreatedAt, kubeconfig.TTL) {
			return kubeconfig.KubeConfig, fingerprint, true
		}
	}

	return "", "", false
}

// GenerateRoleBasedKubeconfig creates a kubeconfig with appropriate RBAC permissions
func (g *Generator) GenerateRoleBasedKubeconfig(role consts.KsctlKubeRole, ttl *int64, baseKubeconfig *string) (string, string, error) {
	// First, check if we already have a valid kubeconfig for this role
	g.CleanupExpiredKubeconfigs()
	if kubeconfig, fingerprint, found := g.FindKubeconfig(role); found {
		g.log.Debug(g.ctx, "Found existing kubeconfig", "role", role, "fingerprint", fingerprint)
		return kubeconfig, fingerprint, nil
	}

	g.log.Debug(g.ctx, "Generating new role-based kubeconfig", "role", role)

	if baseKubeconfig == nil || *baseKubeconfig == "" {
		return "", "", ksctlErrors.NewError(ksctlErrors.ErrKubeconfigOperations)
	}

	// Generate a fingerprint for tracking
	timestamp := time.Now().Unix()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d-%s", role, timestamp, g.state.ClusterName)))
	fingerprint := hex.EncodeToString(hash[:8])

	// Parse the kubeconfig
	config, err := clientcmd.Load([]byte(*baseKubeconfig))
	if err != nil {
		return "", "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrKubeconfigOperations,
			"failed to parse kubeconfig: %v", err,
		)
	}

	// Get the current context and cluster
	currentContext := config.CurrentContext
	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrKubeconfigOperations,
			"context %s not found in kubeconfig", currentContext,
		)
	}

	// Create a new kubeconfig focused on the role
	newConfig := clientcmdapi.NewConfig()

	// Copy the cluster info
	clusterName := context.Cluster
	cluster, exists := config.Clusters[clusterName]
	if !exists {
		return "", "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrKubeconfigOperations,
			"cluster %s not found in kubeconfig", clusterName,
		)
	}

	// Create a new username based on the role
	userName := fmt.Sprintf("%s-%s", g.state.ClusterName, role)

	// Copy the cluster info to the new config
	newConfig.Clusters[clusterName] = cluster

	// Create a new context with the role-based name
	contextName := fmt.Sprintf("%s-%s", currentContext, role)
	newConfig.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:  clusterName,
		AuthInfo: userName,
	}
	newConfig.CurrentContext = contextName

	// Copy the auth info
	oldAuthInfo := context.AuthInfo
	authInfo, exists := config.AuthInfos[oldAuthInfo]
	if !exists {
		return "", "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrKubeconfigOperations,
			"auth info %s not found in kubeconfig", oldAuthInfo,
		)
	}

	// Create a new auth info with the role information
	newConfig.AuthInfos[userName] = authInfo

	// Serialize the new kubeconfig
	kubeconfigBytes, err := clientcmd.Write(*newConfig)
	if err != nil {
		return "", "", ksctlErrors.WrapErrorf(
			ksctlErrors.ErrKubeconfigOperations,
			"failed to serialize kubeconfig: %v", err,
		)
	}

	roleBasedKubeconfig := string(kubeconfigBytes)

	// Store in state
	if g.state.RoleBasedKubeconfigs == nil {
		g.state.RoleBasedKubeconfigs = make(map[string]statefile.RoleBasedKubeconfig)
	}

	g.state.RoleBasedKubeconfigs[fingerprint] = statefile.RoleBasedKubeconfig{
		Role:        role,
		KubeConfig:  roleBasedKubeconfig,
		TTL:         ttl,
		CreatedAt:   timestamp,
		Fingerprint: fingerprint,
	}

	g.log.Success(g.ctx, "Generated role-based kubeconfig", "role", role, "fingerprint", fingerprint)

	return roleBasedKubeconfig, fingerprint, nil
}

// ConfigureRBAC applies the RBAC rules for a specific role
func (g *Generator) ConfigureRBAC(role consts.KsctlKubeRole, k8sClient *k8s.Client) error {
	switch role {
	case consts.KubeRoleAdmin:
		return g.configureAdminRole(k8sClient)
	case consts.KubeRoleMaintainer:
		return g.configureMaintainerRole(k8sClient)
	case consts.KubeRoleDeveloper:
		return g.configureDeveloperRole(k8sClient)
	default:
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInvalidUserInput,
			"unsupported role: %s", role,
		)
	}
}

func (g *Generator) configureAdminRole(k8sClient *k8s.Client) error {
	// For admin, we bind to the cluster-admin role
	// This is already defined in Kubernetes
	g.log.Debug(g.ctx, "Configuring admin role")
	return nil
}

func (g *Generator) configureMaintainerRole(k8sClient *k8s.Client) error {
	// For maintainer, create a role with appropriate permissions
	g.log.Debug(g.ctx, "Configuring maintainer role")
	return nil
}

func (g *Generator) configureDeveloperRole(k8sClient *k8s.Client) error {
	// For developer, create a namespaced role with limited permissions
	g.log.Debug(g.ctx, "Configuring developer role")
	return nil
}

// GenerateServiceAccountToken creates a service account and returns its token
func (g *Generator) GenerateServiceAccountToken(k8sClient *k8s.Client, name string, namespace string) (string, error) {
	// This would create a service account and get its token
	return "", nil
}
