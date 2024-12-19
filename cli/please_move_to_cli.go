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
