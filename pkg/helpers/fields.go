package helpers

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"golang.org/x/term"
)

func UserInputCredentials(logging resources.LoggerFactory) (string, error) {

	logging.Print("Enter Secret-> ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		logging.Error("Empty secret passed!")
		return UserInputCredentials(logging)
	}
	logging.Print("\n")
	return strings.TrimSpace(string(bytePassword)), nil
}

func ValidateDistro(distro consts.KsctlKubernetes) bool {
	if b := utf8.ValidString(string(distro)); !b {
		return false
	}

	switch distro {
	case consts.K8sK3s, consts.K8sKubeadm, "":
		return true
	default:
		return false
	}
}

func ValidateStorage(storage consts.KsctlStore) bool {
	if b := utf8.ValidString(string(storage)); !b {
		return false
	}

	switch storage {
	case consts.StoreExtMongo, consts.StoreLocal:
		return true
	default:
		return false
	}
}

func ValidCNIPlugin(cni consts.KsctlValidCNIPlugin) bool {

	if b := utf8.ValidString(string(cni)); !b {
		return false
	}

	switch cni {
	case consts.CNIAzure, consts.CNICilium, consts.CNIFlannel, consts.CNIKubenet, consts.CNIKind, "":
		return true
	default:
		return false
	}
}

func ValidateCloud(cloud consts.KsctlCloud) bool {
	if b := utf8.ValidString(string(cloud)); !b {
		return false
	}

	switch cloud {
	case consts.CloudAzure, consts.CloudAws, consts.CloudLocal, consts.CloudAll, consts.CloudCivo:
		return true
	default:
		return false
	}
}

func IsValidName(clusterName string) error {
	if len(clusterName) > 50 {
		return fmt.Errorf("name is too long\tname: %s", clusterName)
	}
	matched, err := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, clusterName)

	if !matched || err != nil {
		return fmt.Errorf("CLUSTER NAME INVALID")
	}

	return nil
}
