package helpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
)

func genOSKubeConfigPath(ctx context.Context) (string, error) {

	var userLoc string
	if v, ok := IsContextPresent(ctx, consts.KsctlCustomDirLoc); ok {
		userLoc = path.Join(strings.Split(strings.TrimSpace(v), " ")...)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userLoc = home
	}

	pathArr := []string{userLoc, ".ksctl", "kubeconfig"}

	return path.Join(pathArr...), nil
}

func WriteKubeConfig(ctx context.Context, kubeconfig string) (string, error) {
	path, err := genOSKubeConfigPath(ctx)
	if err != nil {
		return "", ksctlErrors.ErrInternal.Wrap(err)
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0755); err != nil {
		return "", ksctlErrors.ErrKubeconfigOperations.Wrap(err)
	}

	return path, nil
}

func GenRandomString(length int) (string, error) {
	const letters string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", ksctlErrors.ErrUnknown.Wrap(err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func GenerateInitScriptForVM(resName string) (string, error) {

	postfixStr, err := GenRandomString(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`#!/bin/bash
sudo hostname %s-%s

sudo cp /etc/localtime /etc/localtime.backup

sudo ln -sf /usr/share/zoneinfo/UTC /etc/localtime

`, resName, postfixStr), nil
}
