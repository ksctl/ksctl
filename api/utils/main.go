// ////////////// THIS FILE IS NOT READY ////////////////
package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/kubesimplify/ksctl/api/resources"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kubesimplify/ksctl/api/logger"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Machine struct {
	ManagedNodes        int    `json:"managed_nodes"`
	MachineType         string `json:"machine_type"`
	HAControlPlaneNodes int    `json:"no_cp"`
	HAWorkerNodes       int    `json:"no_wp"`
}

type SSHPayload struct {
	UserName       string `json:"user_name"`
	PathPrivateKey string `json:"path_private_key"`
	PublicIP       string `json:"public_ip"`
	Output         string `json:"output"`
}

type SSHCollection interface {
	SSHExecute(int, *string, bool)
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

//
//type PrinterKubeconfigPATH interface {
//	Printer(logger.Logger, bool, int)
//}
//

func IsValidName(clusterName string) error {
	matched, err := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, clusterName)

	if !matched || err != nil {
		return fmt.Errorf("CLUSTER NAME INVALID")
	}

	return nil
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

//
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

func SaveCred(storage resources.StateManagementInfrastructure, config interface{}, provider string) error {

	if strings.Compare(provider, "civo") != 0 &&
		strings.Compare(provider, "azure") != 0 &&
		strings.Compare(provider, "aws") != 0 {
		return fmt.Errorf("invalid provider (given): Unable to save configuration")
	}

	storeBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = storage.Permission(0640).Path(GetPath(CREDENTIAL_PATH, provider)).Save(storeBytes)
	if err != nil {
		return err
	}

	storage.Logger().Info("[secrets] configuration")
	return nil
}

func GetCred(storage resources.StateManagementInfrastructure, provider string) (i map[string]string, err error) {

	fileBytes, err := storage.Path(GetPath(CREDENTIAL_PATH, provider)).Load()
	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &i)

	if err != nil {
		return
	}
	storage.Logger().Info("🔄 configuration")

	return
}

//func SaveState(logging logger.Logger, config interface{}, provider, clusterType string, clusterDir string) error {
//	if strings.Compare(provider, "civo") != 0 &&
//		strings.Compare(provider, "azure") != 0 &&
//		strings.Compare(provider, "aws") != 0 {
//		return fmt.Errorf("invalid Provider (given): Unable to save configuration")
//	}
//	storeBytes, err := json.Marshal(config)
//	if err != nil {
//		return err
//	}
//	if err := os.MkdirAll(GetPath(CLUSTER_PATH, provider, clusterType, clusterDir), 0755); err != nil && !os.IsExist(err) {
//		return err
//	}
//	_, err = os.Create(GetPath(CLUSTER_PATH, provider, clusterType, clusterDir, "info.json"))
//	if err != nil && !os.IsExist(err) {
//		return err
//	}
//	err = os.WriteFile(GetPath(CLUSTER_PATH, provider, clusterType, clusterDir, "info.json"), storeBytes, 0640)
//	if err != nil {
//		return err
//	}
//	logging.Info("💾 configuration", "")
//	return nil
//}

//func GetState(logging logger.Logger, provider, clusterType, clusterDir string) (i map[string]interface{}, err error) {
//	fileBytes, err := os.ReadFile(GetPath(CLUSTER_PATH, provider, clusterType, clusterDir, "info.json"))
//
//	if err != nil {
//		return
//	}
//
//	err = json.Unmarshal(fileBytes, &i)
//
//	if err != nil {
//		return
//	}
//	logging.Info("🔄 configuration", "")
//	return
//}

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

func CreateSSHKeyPair(provider, clusterDir string) (string, error) {

	pathTillFolder := ""
	pathTillFolder = getPaths(provider, "ha", clusterDir)

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", "keypair")
	cmd.Dir = pathTillFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fmt.Println(string(out))

	keyPairToUpload := GetPath(OTHER_PATH, provider, "ha", clusterDir, "keypair.pub")
	fileBytePub, err := os.ReadFile(keyPairToUpload)
	if err != nil {
		return "", err
	}

	return string(fileBytePub), nil
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

func returnServerPublicKeys(publicIP string) (string, error) {
	c1 := exec.Command("ssh-keyscan", "-t", "rsa", publicIP)
	c2 := exec.Command("ssh-keygen", "-lf", "-")

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	err := c1.Start()
	if err != nil {
		return "", nil
	}
	err = c2.Start()
	if err != nil {
		return "", nil
	}
	err = c1.Wait()
	if err != nil {
		return "", nil
	}
	err = w.Close()
	if err != nil {
		return "", nil
	}
	err = c2.Wait()
	if err != nil {
		return "", nil
	}

	ret := b2.String()

	ret = strings.TrimSpace(ret)

	fingerprint := strings.Split(ret, " ")

	return fingerprint[1], nil
}

func (sshPayload *SSHPayload) SSHExecute(logging logger.Logger, flag int, script string, fastMode bool) error {

	// TODO: Try to use the state Get()
	privateKeyBytes, err := os.ReadFile(sshPayload.PathPrivateKey)
	if err != nil {
		return err
	}

	// create signer
	signer, err := signerFromPem(privateKeyBytes)
	if err != nil {
		return err
	}
	logging.Info("SSH into", fmt.Sprintf("%s@%s", sshPayload.UserName, sshPayload.PublicIP))
	config := &ssh.ClientConfig{
		User: sshPayload.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},

		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSASHA256,
		},
		HostKeyCallback: ssh.HostKeyCallback(
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				actualFingerprint := ssh.FingerprintSHA256(key)
				keyType := key.Type()
				if keyType == ssh.KeyAlgoRSA {
					expectedFingerprint, err := returnServerPublicKeys(sshPayload.PublicIP)
					if err != nil {
						return err
					}
					if expectedFingerprint != actualFingerprint {
						return fmt.Errorf("mismatch of fingerprint")
					}
					return nil
				}
				return fmt.Errorf("unsupported key type: %s", keyType)
			})}

	if !fastMode {
		time.Sleep(SSH_PAUSE_IN_SECONDS * time.Second)
	}

	var conn *ssh.Client
	currRetryCounter := 0

	for currRetryCounter < MAX_RETRY_COUNT {
		conn, err = ssh.Dial("tcp", sshPayload.PublicIP+":22", config)
		if err == nil {
			break
		} else {
			logging.Err(fmt.Sprintln("RETRYING", err))
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == MAX_RETRY_COUNT {
		return fmt.Errorf("🚨 💀 COULDN'T RETRY: %v", err)
	}

	logging.Info("🤖 Exec Scripts", "")
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
	err = session.Run(script)
	if flag == EXEC_WITH_OUTPUT {
		sshPayload.Output = buff.String()
	}
	if err != nil {
		return err
	}

	return nil
}

func UserInputCredentials(logging logger.LogFactory) (string, error) {

	fmt.Print("    Enter Secret-> ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		logging.Err("Empty secret passed!")
		return UserInputCredentials(logging)
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

func IsValidNoOfControlPlanes(noCP int) error {
	if noCP < 3 || (noCP)&1 == 0 {
		return fmt.Errorf("no of controlplanes must be >= 3 and should be odd number")
	}
	return nil
}
