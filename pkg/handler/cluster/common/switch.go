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
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/provider/aws"
	"github.com/ksctl/ksctl/pkg/provider/azure"
	"github.com/ksctl/ksctl/pkg/provider/local"
)

func (kc *Controller) Switch() (*string, error) {
	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	clusterType := consts.ClusterTypeMang
	if kc.b.IsHA(kc.p) {
		clusterType = consts.ClusterTypeHa
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		clusterType); err != nil {

		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			kc.l.Error("StorageClass Kill failed", "reason", err)
		}
	}()

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, aws.ProvideClient)
		if err != nil {
			break
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, local.ProvideClient)

	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if errInit := kc.p.Cloud.InitState(consts.OperationGet); errInit != nil {
		kc.l.Error("handled error", "catch", errInit)
		return nil, errInit
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	kubeconfig, err := kc.p.Cloud.GetKubeconfig()
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	if kubeconfig == nil {
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			kc.l.NewError(
				kc.ctx, "Problem in kubeconfig get"),
		)

		kc.l.Error("Kubeconfig we got is nil")
		return nil, err
	}

	path, err := writeKubeConfig(kc.ctx, *kubeconfig)
	kc.l.Debug(kc.ctx, "data", "kubeconfigPath", path)

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return nil, err
	}

	kc.printKubeConfig(path)

	return kubeconfig, nil
}

func genOSKubeConfigPath(ctx context.Context) (string, error) {

	var userLoc string
	if v, ok := config.IsContextPresent(ctx, consts.KsctlCustomDirLoc); ok {
		userLoc = filepath.Join(strings.Split(strings.TrimSpace(v), " ")...)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userLoc = home
	}

	pathArr := []string{userLoc, ".ksctl", "kubeconfig"}

	return filepath.Join(pathArr...), nil
}

func writeKubeConfig(ctx context.Context, kubeconfig string) (string, error) {
	path, err := genOSKubeConfigPath(ctx)
	if err != nil {
		return "", ksctlErrors.WrapError(ksctlErrors.ErrInternal, err)
	}

	dir, _ := filepath.Split(path)

	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", ksctlErrors.WrapError(ksctlErrors.ErrKubeconfigOperations, err)
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0755); err != nil {
		return "", ksctlErrors.WrapError(ksctlErrors.ErrKubeconfigOperations, err)
	}

	return path, nil
}
