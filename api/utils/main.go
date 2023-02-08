/*
Kubesimplify
@maintainer: 	Dipankar Das <dipankardas0115@gmail.com>
				Anurag Kumar <contact.anurag7@gmail.com>
				Avinesh Tripathi <avineshtripathi1@gmail.com>
*/

package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
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

const (
	SSH_PAUSE_IN_SECONDS = 20
	MAX_RETRY_COUNT      = 8
	CREDENTIAL_PATH      = int(0)
	CLUSTER_PATH         = int(1)
	SSH_PATH             = int(2)
	OTHER_PATH           = int(3)
	EXEC_WITH_OUTPUT     = int(1)
	EXEC_WITHOUT_OUTPUT  = int(0)
)

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

// IsValidRegionCIVO validates the region code for CIVO
func IsValidRegionCIVO(reg string) bool {
	return strings.Compare(reg, "FRA1") == 0 ||
		strings.Compare(reg, "NYC1") == 0 ||
		strings.Compare(reg, "PHX1") == 0 ||
		strings.Compare(reg, "LON1") == 0
}

func IsValidName(clusterName string) bool {
	matched, _ := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, clusterName)

	return matched
}

// getKubeconfig returns the path to clusters specific to provider
func getKubeconfig(provider string, params ...string) string {
	if strings.Compare(provider, "civo") != 0 &&
		strings.Compare(provider, "local") != 0 &&
		strings.Compare(provider, "azure") != 0 &&
		strings.Compare(provider, "aws") != 0 {
		return ""
	}
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

// getCredentials generate the path to the credentials of different providers
func getCredentials(provider string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\.ksctl\\cred\\%s.json", GetUserName(), provider)
	} else {
		return fmt.Sprintf("%s/.ksctl/cred/%s.json", GetUserName(), provider)
	}
}

// GetPath use this in every function and differentiate the logic by using if-else
func GetPath(flag int, provider string, subfolders ...string) string {
	switch flag {
	case SSH_PATH:
		return getSSHPath(provider, subfolders...)
	case CLUSTER_PATH:
		return getKubeconfig(provider, subfolders...)
	case CREDENTIAL_PATH:
		return getCredentials(provider)
	case OTHER_PATH:
		return getPaths(provider, subfolders...)
	default:
		return ""
	}
}

// getSSHPath generate the SSH keypair location and subsequent fetch
func getSSHPath(provider string, params ...string) string {
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

// getPaths to generate path irrespective of the cluster
// its a free flowing (Provider field has not much significance)
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

	keyPairToUpload := GetPath(SSH_PATH, provider, "ha", clusterName+" "+region) + ".pub"
	fileBytePub, err := os.ReadFile(keyPairToUpload)
	if err != nil {
		return "", err
	}

	return string(fileBytePub), nil
}

type SSHPayload struct {
	UserName       string
	PathPrivateKey string
	PublicIP       string
	Output         string
}

type SSHCollection interface {
	SSHExecute(int, *string, bool)
}

func signerFromPem(pemBytes []byte) (ssh.Signer, error) {

	// read pem block
	err := errors.New("pem decode failed, no key found")
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, err
	}
	if x509.IsEncryptedPEMBlock(pemBlock) {
		return nil, fmt.Errorf("pem file is encrypted")
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing plain private key failed %v", err)
	}

	return signer, nil
}

func (sshPayload *SSHPayload) SSHExecute(flag int, script string, fastMode bool) error {

	// var err error
	// publicKeyBytes, err := os.ReadFile(sshPayload.PathPrivateKey + ".pub")
	// if err != nil {
	// 	return err
	// }
	// publicKey, err := ssh.ParsePublicKey(publicKeyBytes)
	// if err != nil {
	// 	return err
	// }

	privateKeyBytes, err := os.ReadFile(sshPayload.PathPrivateKey)
	if err != nil {
		return err
	}

	// create signer
	signer, err := signerFromPem(privateKeyBytes)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// FIXME: Remove the InsecureIgnoreHostKey
		HostKeyCallback: ssh.HostKeyCallback(
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				fmt.Println(key)
				// check the fingerprint of hostkey and server key
				// fmt.Println(publicKey)

				return nil
			}),
	}

	if !fastMode {
		time.Sleep(SSH_PAUSE_IN_SECONDS * time.Second)
	}

	// var err error
	var conn *ssh.Client
	currRetryCounter := 0

	for currRetryCounter < MAX_RETRY_COUNT {
		conn, err = ssh.Dial("tcp", sshPayload.PublicIP+":22", config)
		if err == nil {
			break
		} else {
			log.Printf("❗ RETRYING %v\n", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return fmt.Errorf("🚨 💀 COULDN'T RETRY: %v", err)
	}

	log.Println("🤖 Exec Scripts")
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()
	var buff bytes.Buffer

	sshPayload.Output = ""
	if flag == EXEC_WITH_OUTPUT {
		session.Stdout = &buff
	}
	if err := session.Run(script); err != nil {
		return err
	}
	if flag == EXEC_WITH_OUTPUT {
		sshPayload.Output = buff.String()
	}

	return nil
}
