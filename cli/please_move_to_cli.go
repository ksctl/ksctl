// Copyright 2024 ksctl
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

package cli

import (
	"context"
	"github.com/ksctl/ksctl/pkg/consts"
	"os"
	"path/filepath"
	"strings"
)

func genOSKubeConfigPath(ctx context.Context) (string, error) {

	var userLoc string
	if v, ok := IsContextPresent(ctx, consts.KsctlCustomDirLoc); ok {
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

func WriteKubeConfig(ctx context.Context, kubeconfig string) (string, error) {
	path, err := genOSKubeConfigPath(ctx)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(err)
	}

	dir, _ := filepath.Split(path)

	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", ksctlErrors.ErrKubeconfigOperations.Wrap(err)
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0755); err != nil {
		return "", ksctlErrors.ErrKubeconfigOperations.Wrap(err)
	}

	return path, nil
}
