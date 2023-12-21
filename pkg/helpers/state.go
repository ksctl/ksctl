package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/resources"

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
)

var (
	KSCTL_CONFIG_DIR = func() string {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("%s\\.ksctl", GetUserName())
		}
		return fmt.Sprintf("%s/.ksctl", GetUserName())
	}()
)

// GetUserName returns current active username
func GetUserName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}

// getKubeconfig returns the path to clusters specific to provider

func getKubeconfig(provider consts.KsctlCloud, clusterType consts.KsctlClusterType, params ...string) string {
	if provider != consts.CloudCivo &&
		provider != consts.CloudLocal &&
		provider != consts.CloudAzure &&
		provider != consts.CloudAws {
		return ""
	}
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\config\\%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("\\" + string(clusterType))

		for _, item := range params {
			ret.WriteString("\\" + item)
		}
	} else {
		ret.WriteString(fmt.Sprintf("%s/config/%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("/" + string(clusterType))

		for _, item := range params {
			ret.WriteString("/" + item)
		}
	}
	return ret.String()
}

// getCredentials generate the path to the credentials of different providers

func getCredentials(provider consts.KsctlCloud) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\cred\\%s.json", KSCTL_CONFIG_DIR, provider)
	} else {
		return fmt.Sprintf("%s/cred/%s.json", KSCTL_CONFIG_DIR, provider)
	}
}

//
// GetPath use this in every function and differentiate the logic by using if-else

// make getPath use 3 predefined const last is clusterType TODO:
func GetPath(flag consts.KsctlUtilsConsts, provider consts.KsctlCloud, clusterType consts.KsctlClusterType, subfolders ...string) string {
	// for using different KSCTL DIRECTORY
	if dirName := os.Getenv(string(consts.KsctlCustomDirEnabled)); len(dirName) != 0 {
		KSCTL_CONFIG_DIR = dirName
	}
	switch flag {
	case consts.UtilSSHPath:
		return getSSHPath(provider, clusterType, subfolders...)
	case consts.UtilClusterPath:
		return getKubeconfig(provider, clusterType, subfolders...)
	case consts.UtilCredentialPath:
		return getCredentials(provider)
	case consts.UtilOtherPath:
		return getPaths(provider, clusterType, subfolders...)
	default:
		return ""
	}
}

func SaveCred(storage resources.StorageFactory, log resources.LoggerFactory, config interface{}, provider consts.KsctlCloud) error {

	if provider != consts.CloudCivo && provider != consts.CloudAzure {
		return log.NewError("invalid provider (given): Unable to save configuration")
	}

	storeBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = storage.Permission(0640).Path(GetPath(consts.UtilCredentialPath, provider, "")).Save(storeBytes)
	if err != nil {
		return err
	}

	log.Success("successful in saving credentials")
	return nil
}

func GetCred(storage resources.StorageFactory, log resources.LoggerFactory, provider consts.KsctlCloud) (i map[string]string, err error) {

	fileBytes, err := storage.Path(GetPath(consts.UtilCredentialPath, provider, "")).Load()
	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &i)

	if err != nil {
		return
	}
	log.Success("successful in fetching credentials")

	return
}

// getSSHPath generate the SSH keypair location and subsequent fetch
func getSSHPath(provider consts.KsctlCloud, clusterType consts.KsctlClusterType, params ...string) string {
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\config\\%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("\\" + string(clusterType))

		for _, item := range params {
			ret.WriteString("\\" + item)
		}
		ret.WriteString("\\keypair")
	} else {
		ret.WriteString(fmt.Sprintf("%s/config/%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("/" + string(clusterType))

		for _, item := range params {
			ret.WriteString("/" + item)
		}
		ret.WriteString("/keypair")
	}
	return ret.String()
}

// getPaths to generate path irrespective of the cluster
// its a free flowing (Provider field has not much significance)
func getPaths(provider consts.KsctlCloud, clusterType consts.KsctlClusterType, params ...string) string {
	var ret strings.Builder
	if dirName := os.Getenv(string(consts.KsctlCustomDirEnabled)); len(dirName) != 0 {
		KSCTL_CONFIG_DIR = dirName
	}

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\config\\%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("\\" + string(clusterType))

		for _, item := range params {
			ret.WriteString("\\" + item)
		}
	} else {
		ret.WriteString(fmt.Sprintf("%s/config/%s", KSCTL_CONFIG_DIR, provider))
		ret.WriteString("/" + string(clusterType))

		for _, item := range params {
			ret.WriteString("/" + item)
		}
	}
	return ret.String()
}
