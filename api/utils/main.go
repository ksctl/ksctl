/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type AwsProvider struct {
	ClusterName string
	HACluster   bool
	Region      string
	Spec        Machine
	AccessKey   string
	Secret      string
}

type AzureProvider struct {
	ClusterName         string
	HACluster           bool
	Region              string
	Spec                Machine
	SubscriptionID      string
	TenantID            string
	ServicePrincipleKey string
	ServicePrincipleID  string
}

type Machine struct {
	ManagedNodes        int
	Disk                string
	HAControlPlaneNodes int
	HAWorkerNodes       int
	Mem                 string
	Cpu                 string
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Spec        Machine
}

type CivoCredential struct {
	Token string `json:"token"`
}

type AzureCredential struct {
	SubscriptionID      string `json:"subscription_id"`
	TenantID            string `json:"tenant_id"`
	ServicePrincipleKey string `json:"service_principal_key"`
	ServicePrincipleID  string `json:"service_principal_id"`
}

type AwsCredential struct {
	AccesskeyID string `json:"access_key_id"`
	Secret      string `json:"secret_access_key"`
}

// GetUserName returns current active username
func GetUserName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}

type PrinterKubeconfigPATH interface {
	Printer(bool, int)
}

type CivoHandlers interface {
	CreateCluster() error
	DeleteCluster() error
	SwitchContext() error
	AddMoreWorkerNodes() error
	DeleteSomeWorkerNodes() error
}

func IsValidRegionCIVO(reg string) bool {
	return strings.Compare(reg, "FRA1") == 0 ||
		strings.Compare(reg, "NYC1") == 0 ||
		strings.Compare(reg, "LON1") == 0
}

func helperASCII(character uint8) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

func helperDIGIT(character uint8) bool {
	return character >= '0' && character <= '9'
}

func helperSPECIAL(character uint8) bool {
	return character == '-' || character == '_'
}

// TODO: Use Regex expression for valid clusterNames
func IsValidName(clusterName string) bool {

	if !helperASCII(clusterName[0]) &&
		(helperDIGIT(clusterName[0]) || !helperDIGIT(clusterName[0])) {
		return false
	}

	for _, chara := range clusterName {
		if helperASCII(uint8(chara)) || helperDIGIT(uint8(chara)) || helperSPECIAL(uint8(chara)) {
			continue
		} else {
			return false
		}
	}
	return true
}

func GetKubeconfig(provider string, params ...string) string {
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\.ksctl\\config\\%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("\\" + item)
		}
	} else {
		ret.WriteString(fmt.Sprintf("%s/.ksctl/config/%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("/" + item)
		}
	}
	return ret.String()
}

func getCredentials(provider string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\%s.json", GetUserName(), provider)
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/%s.json", GetUserName(), provider)
	}
}

const (
	CREDENTIAL_PATH     = int(0)
	CLUSTER_PATH        = int(1)
	SSH_PATH            = int(2)
	OTHER_PATH          = int(3)
	EXEC_WITH_OUTPUT    = int(1)
	EXEC_WITHOUT_OUTPUT = int(0)
)

// GetPath use this in every function and differentiate the logic by using if-else
// flag is used to indicate 1 -> KUBECONFIG, 0 -> CREDENTIALS
func GetPath(flag int, provider string, subfolders ...string) string {
	switch flag {
	case SSH_PATH:
		return GetSSHPath(provider, subfolders...)
	case CLUSTER_PATH:
		return GetKubeconfig(provider, subfolders...)
	case CREDENTIAL_PATH:
		return getCredentials(provider)
	case OTHER_PATH:
		return getPaths(provider, subfolders...)
	default:
		return ""
	}
}

func GetSSHPath(provider string, params ...string) string {
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\.ksctl\\config\\%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("\\" + item)
		}
		ret.WriteString("\\keypair")
	} else {
		ret.WriteString(fmt.Sprintf("%s/.ksctl/config/%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("/" + item)
		}
		ret.WriteString("/keypair")
	}
	return ret.String()
}

func getPaths(provider string, params ...string) string {
	var ret strings.Builder

	if runtime.GOOS == "windows" {
		ret.WriteString(fmt.Sprintf("%s\\.ksctl\\config\\%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("\\" + item)
		}
	} else {
		ret.WriteString(fmt.Sprintf("%s/.ksctl/config/%s", GetUserName(), provider))
		for _, item := range params {
			ret.WriteString("/" + item)
		}
	}
	return ret.String()
}

// CreateSSHKeyPair return public key and error
func CreateSSHKeyPair(provider, clusterName, region string) (string, error) {

	pathTillFolder := getPaths(provider, "ha", clusterName+" "+region)

	cmd := exec.Command("ssh-keygen", "-N", "", "-f", "keypair")
	cmd.Dir = pathTillFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fmt.Println(string(out))

	keyPairToUpload := GetPath(SSH_PATH, "civo", "ha", clusterName+" "+region) + ".pub"
	fileBytePub, err := os.ReadFile(keyPairToUpload)
	if err != nil {
		return "", err
	}

	return string(fileBytePub), nil
}

type SSHPayload struct {
	UserName       string
	PathPrivateKey string
}

func (sshPayload *SSHPayload) SSHExecute(flag int, output *string) {
}
