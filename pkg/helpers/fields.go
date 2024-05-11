package helpers

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ksctl/ksctl/internal/storage/types"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"golang.org/x/term"
)

func UserInputCredentials(ctx context.Context, logging resources.LoggerFactory) (string, error) {

	logging.Print(ctx, "Enter Secret")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		logging.Error(ctx, "Empty secret passed!")
		return UserInputCredentials(ctx, logging)
	}
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
	case consts.StoreExtMongo, consts.StoreLocal, consts.StoreK8s:
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

func IsValidVersion(ver string) error {
	if ver == "latest" || ver == "stable" {
		return nil
	}

	patternWithoutVPrefix := `^\d+(\.\d{1,2}){0,2}$`
	patternWithVPrefix := `^v\d+(\.\d{1,2}){0,2}$`
	matchStringWithoutVPrefix, err := regexp.MatchString(patternWithoutVPrefix, ver)
	if err != nil {
		return fmt.Errorf("failed to compile argo-rollouts regex. %v", err)
	}
	matchStringWithVPrefix, err := regexp.MatchString(patternWithVPrefix, ver)
	if err != nil {
		return fmt.Errorf("failed to compile argo-rollouts regex. %v", err)
	}

	if !matchStringWithoutVPrefix && !matchStringWithVPrefix {
		return fmt.Errorf("version `%s` is not valid", ver)
	}
	return nil
}

func ToApplicationTempl(apps []string) ([]types.Application, error) {

	_apps := make([]types.Application, 0)
	for _, app := range apps {

		temp := strings.Split(app, "@")

		if len(temp) > 2 || len(app) == 0 {
			return nil, fmt.Errorf("invalid format for application should be APP_NAME@VERSION")
		}
		if len(temp) == 1 {
			// version was not specified
			_apps = append(_apps, types.Application{
				Name:    temp[0],
				Version: "latest",
			})
		} else {
			_apps = append(_apps, types.Application{
				Name:    temp[0],
				Version: temp[1],
			})
		}
	}
	return _apps, nil
}
