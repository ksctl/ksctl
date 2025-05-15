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

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	bootstrapHandler "github.com/ksctl/ksctl/v2/pkg/bootstrap/handler"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/kubeconfig"
	"github.com/ksctl/ksctl/v2/pkg/provider/aws"
	"github.com/ksctl/ksctl/v2/pkg/provider/azure"
	"github.com/ksctl/ksctl/v2/pkg/provider/local"
)

// KubeconfigResponse provides the kubeconfig information
type KubeconfigResponse struct {
	KubeConfig  string
	Role        consts.KsctlKubeRole
	Fingerprint string
	ExpiresAt   *string // Human-readable expiration time, nil if no TTL
	SavedTo     *string // Path where kubeconfig was saved, nil if not saved to file
}

// ListKubeconfigs lists all role-based kubeconfigs and returns their details
func (kc *Controller) ListKubeconfigs() ([]KubeconfigResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			kc.l.Error("panic recovered", "error", err)
		}
	}()

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		kc.p.Metadata.ClusterType); err != nil {
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("failed to kill storage", "error", err)
		}
	}()

	// Initialize cloud provider
	if err := kc.initCloudProvider(); err != nil {
		return nil, err
	}

	if err := kc.p.Cloud.InitState(consts.OperationGet); err != nil {
		return nil, err
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		return nil, err
	}

	generator := kubeconfig.NewGenerator(kc.ctx, kc.l, kc.s)
	generator.CleanupExpiredKubeconfigs()

	if kc.s.RoleBasedKubeconfigs == nil || len(kc.s.RoleBasedKubeconfigs) == 0 {
		return []KubeconfigResponse{}, nil
	}

	var response []KubeconfigResponse
	for _, rbkc := range kc.s.RoleBasedKubeconfigs {
		var expiresAt *string
		if rbkc.TTL != nil {
			expiryTime := time.Unix(rbkc.CreatedAt+*rbkc.TTL, 0).Format(time.RFC3339)
			expiresAt = &expiryTime
		}

		response = append(response, KubeconfigResponse{
			KubeConfig:  rbkc.KubeConfig,
			Role:        rbkc.Role,
			Fingerprint: rbkc.Fingerprint,
			ExpiresAt:   expiresAt,
			SavedTo:     nil,
		})
	}

	return response, nil
}

// GetKubeconfig retrieves a kubeconfig for a specific role
func (kc *Controller) GetKubeconfig(role consts.KsctlKubeRole, ttl *int64, outputPath string) (*KubeconfigResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			kc.l.Error("panic recovered", "error", err)
		}
	}()

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		kc.p.Metadata.ClusterType); err != nil {
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("failed to kill storage", "error", err)
		}
	}()

	// Initialize cloud provider
	if err := kc.initCloudProvider(); err != nil {
		return nil, err
	}

	if err := kc.p.Cloud.InitState(consts.OperationGet); err != nil {
		return nil, err
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		return nil, err
	}

	// Get the current kubeconfig as a base
	baseKubeconfig, err := kc.p.Cloud.GetKubeconfig()
	if err != nil {
		return nil, err
	}

	if baseKubeconfig == nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			kc.l.NewError(kc.ctx, "kubeconfig not found for cluster"),
		)
	}

	// Create a k8s client to configure RBAC if needed
	k8sClient, err := bootstrapHandler.NewClusterClient(
		kc.ctx,
		kc.l,
		kc.p.Storage,
		*baseKubeconfig,
	)
	if err != nil {
		kc.l.Warn(kc.ctx, "Could not create k8s client. RBAC configuration skipped.", "error", err)
		// Continue without RBAC configuration
	}

	// Generate role-based kubeconfig
	generator := kubeconfig.NewGenerator(kc.ctx, kc.l, kc.s)

	// Configure RBAC if we have a client
	if k8sClient != nil {
		if err := generator.ConfigureRBAC(role, k8sClient.GetK8sClient()); err != nil {
			kc.l.Warn(kc.ctx, "Failed to configure RBAC", "role", role, "error", err)
			// Continue anyway - the kubeconfig will use the base permissions
		}
	}

	// Generate the role-based kubeconfig
	roleBasedConfig, fingerprint, err := generator.GenerateRoleBasedKubeconfig(role, ttl, baseKubeconfig)
	if err != nil {
		return nil, err
	}

	// Save to state
	if err := kc.p.Storage.Write(kc.s); err != nil {
		kc.l.Warn(kc.ctx, "Failed to save role-based kubeconfig to state", "error", err)
		// Continue anyway - we still have the kubeconfig
	}

	// Build the response
	response := &KubeconfigResponse{
		KubeConfig:  roleBasedConfig,
		Role:        role,
		Fingerprint: fingerprint,
		SavedTo:     nil,
	}

	// Add expiry time if TTL is set
	if ttl != nil {
		expiryTime := time.Now().Add(time.Duration(*ttl) * time.Second).Format(time.RFC3339)
		response.ExpiresAt = &expiryTime
	}

	// Write to file if requested
	if outputPath != "" {
		// Create directories if they don't exist
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return response, ksctlErrors.WrapErrorf(
				ksctlErrors.ErrInternal,
				"failed to create directories for kubeconfig: %v", err,
			)
		}

		// Write kubeconfig to file
		if err := os.WriteFile(outputPath, []byte(roleBasedConfig), 0600); err != nil {
			return response, ksctlErrors.WrapErrorf(
				ksctlErrors.ErrInternal,
				"failed to write kubeconfig to file: %v", err,
			)
		}

		response.SavedTo = &outputPath
	}

	return response, nil
}

// RevokeKubeconfig revokes a specific kubeconfig by fingerprint
func (kc *Controller) RevokeKubeconfig(fingerprint string) error {
	defer func() {
		if err := recover(); err != nil {
			kc.l.Error("panic recovered", "error", err)
		}
	}()

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		kc.p.Metadata.ClusterType); err != nil {
		return err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("failed to kill storage", "error", err)
		}
	}()

	if err := kc.p.Cloud.InitState(consts.OperationGet); err != nil {
		return err
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		return err
	}

	if kc.s.RoleBasedKubeconfigs == nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			kc.l.NewError(kc.ctx, "no role-based kubeconfigs found"),
		)
	}

	// Check if the fingerprint exists
	if _, exists := kc.s.RoleBasedKubeconfigs[fingerprint]; !exists {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			kc.l.NewError(kc.ctx, "kubeconfig not found for fingerprint", "fingerprint", fingerprint),
		)
	}

	// Delete the kubeconfig
	delete(kc.s.RoleBasedKubeconfigs, fingerprint)

	// Save the state
	if err := kc.p.Storage.Write(kc.s); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			kc.l.NewError(kc.ctx, "failed to save state after removing kubeconfig"),
		)
	}

	kc.l.Success(kc.ctx, "Kubeconfig revoked", "fingerprint", fingerprint)
	return nil
}

// initCloudProvider initializes the cloud provider based on the metadata
func (kc *Controller) initCloudProvider() error {
	var err error

	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, azure.ProvideClient)
		if err != nil {
			return err
		}

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, aws.ProvideClient)
		if err != nil {
			return err
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, local.ProvideClient)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported cloud provider: %s", kc.p.Metadata.Provider)
	}

	err = kc.p.Cloud.InitState(consts.OperationGet)
	if err != nil {
		return err
	}

	return nil
}
