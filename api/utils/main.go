/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package payload

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type Machine struct {
	Nodes int
	Cpu   string
	Mem   string
	Disk  string
}

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
	servicePrincipleKey string
	servicePrincipleID  string
}

type CivoProvider struct {
	ClusterName string
	APIKey      string
	HACluster   bool
	Region      string
	Spec        Machine
	Application string
	CNIPlugin   string
}

type LocalProvider struct {
	ClusterName string
	HACluster   bool
	Spec        Machine
}

//type Providers struct {
//	eks  *AwsProvider
//	aks  *AzureProvider
//	k3s  *CivoProvider
//	mk8s *LocalProvider
//}

// GetUserName returns current active username
func GetUserName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}

type PrinterKubeconfigPATH interface {
	Printer(int)
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

func getKubeconfigCIVO(params ...string) string {
	var ret string

	if runtime.GOOS == "windows" {
		ret = fmt.Sprintf("%s\\.ksctl\\config", GetUserName())
		for _, item := range params {
			ret += "\\" + item
		}
	} else {
		ret = fmt.Sprintf("%s/.ksctl/config", GetUserName())
		for _, item := range params {
			ret += "/" + item
		}
	}
	return ret
}

func getCredentialsCIVO() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\civo", GetUserName())
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/civo", GetUserName())
	}
}

// GetPath use this in every function and differentiate the logic by using if-else
// flag is used to indicate 1 -> KUBECONFIG, 0 -> CREDENTIALS
func GetPathCIVO(flag int8, params ...string) string {
	switch flag {
	case 1:
		return getKubeconfigCIVO(params...)
	case 0:
		return getCredentialsCIVO()
	default:
		return ""
	}
}
