package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kubesimplify/ksctl/pkg/resources"

	"github.com/kubesimplify/ksctl/pkg/logger"
	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type SSHPayload struct {
	UserName       string
	PathPrivateKey string
	PublicIP       string
	Output         string

	flag     KsctlUtilsConsts
	script   string
	fastMode bool
}

type SSHCollection interface {
	SSHExecute(resources.StorageFactory) error
	Flag(KsctlUtilsConsts) SSHCollection
	Script(string) SSHCollection
	FastMode(bool) SSHCollection
	Username(string)
	LocPrivateKey(string)
	GetOutput() string
	IPv4(ip string) SSHCollection
}

var (
	KSCTL_CONFIG_DIR = func() string {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("%s\\.ksctl", GetUserName())
		}
		return fmt.Sprintf("%s/.ksctl", GetUserName())
	}()
)

func (ssh *SSHPayload) Username(s string) {
	ssh.UserName = s
}

func (ssh *SSHPayload) LocPrivateKey(s string) {
	ssh.PathPrivateKey = s
}

func (ssh *SSHPayload) GetOutput() string {
	out := ssh.Output
	ssh.Output = ""
	return out
}

func (ssh *SSHPayload) IPv4(ip string) SSHCollection {
	ssh.PublicIP = ip
	return ssh
}

// GetUserName returns current active username
func GetUserName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}

// getKubeconfig returns the path to clusters specific to provider

func getKubeconfig(provider KsctlCloud, clusterType KsctlClusterType, params ...string) string {
	if provider != CloudCivo &&
		provider != CloudLocal &&
		provider != CloudAzure &&
		provider != CloudAws {
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

func getCredentials(provider KsctlCloud) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\cred\\%s.json", KSCTL_CONFIG_DIR, provider)
	} else {
		return fmt.Sprintf("%s/cred/%s.json", KSCTL_CONFIG_DIR, provider)
	}
}

//
// GetPath use this in every function and differentiate the logic by using if-else

// make getPath use 3 predefined const last is clusterType TODO:
func GetPath(flag KsctlUtilsConsts, provider KsctlCloud, clusterType KsctlClusterType, subfolders ...string) string {
	// for using different KSCTL DIRECTORY
	if dirName := os.Getenv(string(KsctlCustomDirEnabled)); len(dirName) != 0 {
		KSCTL_CONFIG_DIR = dirName
	}
	switch flag {
	case UtilSSHPath:
		return getSSHPath(provider, clusterType, subfolders...)
	case UtilClusterPath:
		return getKubeconfig(provider, clusterType, subfolders...)
	case UtilCredentialPath:
		return getCredentials(provider)
	case UtilOtherPath:
		return getPaths(provider, clusterType, subfolders...)
	default:
		return ""
	}
}

func SaveCred(storage resources.StorageFactory, config interface{}, provider KsctlCloud) error {

	if provider != CloudCivo && provider != CloudAzure {
		return fmt.Errorf("invalid provider (given): Unable to save configuration")
	}

	storeBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = storage.Permission(0640).Path(GetPath(UtilCredentialPath, provider, "")).Save(storeBytes)
	if err != nil {
		return err
	}

	storage.Logger().Success("[secrets] configuration")
	return nil
}

func GetCred(storage resources.StorageFactory, provider KsctlCloud) (i map[string]string, err error) {

	fileBytes, err := storage.Path(GetPath(UtilCredentialPath, provider, "")).Load()
	if err != nil {
		return
	}

	err = json.Unmarshal(fileBytes, &i)

	if err != nil {
		return
	}
	storage.Logger().Success("[utils] configuration")

	return
}

// getSSHPath generate the SSH keypair location and subsequent fetch
func getSSHPath(provider KsctlCloud, clusterType KsctlClusterType, params ...string) string {
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
func getPaths(provider KsctlCloud, clusterType KsctlClusterType, params ...string) string {
	var ret strings.Builder
	if dirName := os.Getenv(string(KsctlCustomDirEnabled)); len(dirName) != 0 {
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

func CreateSSHKeyPair(storage resources.StorageFactory, provider KsctlCloud, clusterDir string) (string, error) {

	pathTillFolder := ""
	pathTillFolder = getPaths(provider, ClusterTypeHa, clusterDir)

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", "keypair") // WARN: it requires the os to have these dependencies
	cmd.Dir = pathTillFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	storage.Logger().Print("[utils]", string(out))

	path := GetPath(UtilOtherPath, provider, ClusterTypeHa, clusterDir, "keypair.pub")
	fileBytePub, err := storage.Path(path).Load()
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

func returnServerPublicKeys(provider KsctlCloud, publicIP string) (string, error) {
	var c1 *exec.Cmd
	var c2 *exec.Cmd

	if provider == CloudAws {
		c1 = exec.Command("ssh-keyscan", "-t", "ecdsa", publicIP)
		c2 = exec.Command("ssh-keygen", "-lf", "-")
	} else {
		c1 = exec.Command("ssh-keyscan", "-t", "rsa", publicIP) // WARN: it requires the os to have these dependencies
		c2 = exec.Command("ssh-keygen", "-lf", "-")             // WARN: it requires the os to have these dependencies
	}

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

func (ssh *SSHPayload) Flag(execMethod KsctlUtilsConsts) SSHCollection {
	if execMethod == UtilExecWithOutput || execMethod == UtilExecWithoutOutput {
		ssh.flag = execMethod
		return ssh
	}
	return nil
}

func (ssh *SSHPayload) Script(s string) SSHCollection {
	ssh.script = s
	return ssh
}

func (ssh *SSHPayload) FastMode(mode bool) SSHCollection {
	ssh.fastMode = mode
	return ssh
}

func (sshPayload *SSHPayload) SSHExecute(storage resources.StorageFactory, provider KsctlCloud) error {

	privateKeyBytes, err := storage.Path(sshPayload.PathPrivateKey).Load()
	if err != nil {
		return err
	}

	// create signer
	signer, err := signerFromPem(privateKeyBytes)
	if err != nil {
		return err
	}
	storage.Logger().Success("[ssh] SSH into", fmt.Sprintf("%s@%s", sshPayload.UserName, sshPayload.PublicIP))

	// NOTE: when the fake environment variable is set //
	if fake := os.Getenv(string(KsctlFakeFlag)); len(fake) != 0 {
		storage.Logger().Success("[ssh] Exec Scripts")
		sshPayload.Output = ""

		if sshPayload.flag == UtilExecWithOutput {
			sshPayload.Output = "random fake"
		}
		return nil
	}

	var keyalgo []string
	if provider == CloudAws {
		keyalgo = []string{
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoED25519,
		}
	} else {
		keyalgo = []string{
			ssh.KeyAlgoRSASHA256,
		}
	}

	config := &ssh.ClientConfig{
		User: sshPayload.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},

		HostKeyAlgorithms: keyalgo,

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
						return fmt.Errorf("[ssh] mismatch of fingerprint")
					}
					return nil
				}
				return fmt.Errorf("[ssh] unsupported key type: %s", keyType)
			})}

	if !sshPayload.fastMode {
		time.Sleep(DurationSSHPause)
	}

	var conn *ssh.Client
	currRetryCounter := KsctlCounterConsts(0)

	for currRetryCounter < CounterMaxRetryCount {
		conn, err = ssh.Dial("tcp", sshPayload.PublicIP+":22", config)
		if err == nil {
			break
		} else {
			storage.Logger().Warn(fmt.Sprintln("RETRYING", err))
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == CounterMaxRetryCount {
		return fmt.Errorf("[ssh] maximum retry count reached for ssh conn %v", err)
	}

	storage.Logger().Success("[ssh] Exec Scripts")
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()
	var buff bytes.Buffer

	sshPayload.Output = ""
	if sshPayload.flag == UtilExecWithOutput {
		session.Stdout = &buff
	}
	err = session.Run(sshPayload.script)
	if sshPayload.flag == UtilExecWithOutput {
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

func ValidateDistro(distro KsctlKubernetes) bool {
	if b := utf8.ValidString(string(distro)); !b {
		return false
	}

	switch distro {
	case K8sK3s, K8sKubeadm, "":
		return true
	default:
		return false
	}
}

func ValidateStorage(storage KsctlStore) bool {
	if b := utf8.ValidString(string(storage)); !b {
		return false
	}

	switch storage {
	case StoreRemote, StoreLocal:
		return true
	default:
		return false
	}
}

func ValidCNIPlugin(cni KsctlValidCNIPlugin) bool {

	if b := utf8.ValidString(string(cni)); !b {
		return false
	}

	switch cni {
	case CNIAzure, CNICilium, CNIFlannel, CNIKubenet, CNIKind, "":
		return true
	default:
		return false
	}
}

func ValidateCloud(cloud KsctlCloud) bool {
	if b := utf8.ValidString(string(cloud)); !b {
		return false
	}

	switch cloud {
	case CloudAzure, CloudAws, CloudLocal, CloudAll, CloudCivo:
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
