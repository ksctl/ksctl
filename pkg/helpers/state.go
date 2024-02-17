package helpers

import (
	"os"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

// GetUserName returns current active username
func GetUserName() string {
	return os.Getenv(UserDir)
}

func genOSKubeConfigPath() string {

	var userLoc string
	if v := os.Getenv(string(consts.KsctlCustomDirEnabled)); len(v) != 0 {
		userLoc = strings.Join(strings.Split(strings.TrimSpace(v), " "), PathSeparator)
	} else {
		userLoc = GetUserName()
	}

	pathArr := []string{userLoc, ".ksctl", "kubeconfig"}

	return strings.Join(pathArr, PathSeparator)
}

func WriteKubeConfig(kubeconfig string) (string, error) {
	path := genOSKubeConfigPath()
	err := os.WriteFile(path, []byte(kubeconfig), 0755)
	if err != nil {
		return "", err
	}

	return path, nil
}
