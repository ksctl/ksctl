package helpers

import (
	"crypto/rand"
	"math/big"
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

// GenRandomString it generates RandomString
func GenRandomString(length int) (string, error) {
	const letters string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
