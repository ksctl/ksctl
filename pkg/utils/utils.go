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

	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type SSHPayload struct {
	UserName       string
	PathPrivateKey string
	PublicIP       string
	Output         string

	flag     consts.KsctlUtilsConsts
	script   string
	fastMode bool
}

type SSHCollection interface {
	SSHExecute(resources.StorageFactory, resources.LoggerFactory) error
	Flag(consts.KsctlUtilsConsts) SSHCollection
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

func getKubeconfig(provider consts.KsctlCloud, clusterType consts.KsctlClusterType, params ...string) string {
	if provider != consts.CloudCivo &&
		provider != consts.CloudLocal &&
		provider != consts.CloudAzure {
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

func CreateSSHKeyPair(storage resources.StorageFactory, log resources.LoggerFactory, provider consts.KsctlCloud, clusterDir string) (string, error) {

	pathTillFolder := ""
	pathTillFolder = getPaths(provider, consts.ClusterTypeHa, clusterDir)

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", "keypair") // WARN: it requires the os to have these dependencies
	cmd.Dir = pathTillFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	log.Debug("Printing", "keypair", string(out))

	path := GetPath(consts.UtilOtherPath, provider, consts.ClusterTypeHa, clusterDir, "keypair.pub")
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

func returnServerPublicKeys(publicIP string) (string, error) {
	c1 := exec.Command("ssh-keyscan", "-t", "rsa", publicIP) // WARN: it requires the os to have these dependencies
	c2 := exec.Command("ssh-keygen", "-lf", "-")             // WARN: it requires the os to have these dependencies

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

func (ssh *SSHPayload) Flag(execMethod consts.KsctlUtilsConsts) SSHCollection {
	if execMethod == consts.UtilExecWithOutput || execMethod == consts.UtilExecWithoutOutput {
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

func (sshPayload *SSHPayload) SSHExecute(storage resources.StorageFactory, log resources.LoggerFactory) error {

	privateKeyBytes, err := storage.Path(sshPayload.PathPrivateKey).Load()
	if err != nil {
		return err
	}

	// create signer
	signer, err := signerFromPem(privateKeyBytes)
	if err != nil {
		return err
	}
	log.Debug("SSH into", "sshAddr", fmt.Sprintf("%s@%s", sshPayload.UserName, sshPayload.PublicIP))

	// NOTE: when the fake environment variable is set //
	if fake := os.Getenv(string(consts.KsctlFakeFlag)); len(fake) != 0 {
		log.Debug("Exec Scripts for fake flag")
		sshPayload.Output = ""

		if sshPayload.flag == consts.UtilExecWithOutput {
			sshPayload.Output = "random fake"
		}
		return nil
	}

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
						return log.NewError("mismatch of SSH fingerprint")
					}
					return nil
				}
				return log.NewError("unsupported key type: %s", keyType)
			})}

	if !sshPayload.fastMode {
		time.Sleep(consts.DurationSSHPause)
	}

	var conn *ssh.Client
	currRetryCounter := consts.KsctlCounterConsts(0)

	for currRetryCounter < consts.CounterMaxRetryCount {
		conn, err = ssh.Dial("tcp", sshPayload.PublicIP+":22", config)
		if err == nil {
			break
		} else {
			log.Warn("RETRYING", err)
		}
		time.Sleep(10 * time.Second) // waiting for ssh to get started
		currRetryCounter++
	}
	if currRetryCounter == consts.CounterMaxRetryCount {
		return log.NewError("maximum retry count reached for ssh conn %v", err)
	}

	log.Debug("Printing", "bashScript", sshPayload.script)
	log.Print("Exec Scripts")
	defer conn.Close()

	session, err := conn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()
	var buff bytes.Buffer

	sshPayload.Output = ""
	//if sshPayload.flag == UtilExecWithOutput {
	session.Stdout = &buff // make the stdout be stored in buffer
	//}
	err = session.Run(sshPayload.script)

	bufferContent := buff.String()
	log.Debug("Printing", "CommandResult", bufferContent)

	if sshPayload.flag == consts.UtilExecWithOutput {
		sshPayload.Output = bufferContent
	}

	if err != nil {
		return err
	}

	log.Success("Success in executing the script")

	return nil
}

func UserInputCredentials(logging resources.LoggerFactory) (string, error) {

	fmt.Print("    Enter Secret-> ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		logging.Error("Empty secret passed!")
		return UserInputCredentials(logging)
	}
	fmt.Println()
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
	case consts.StoreRemote, consts.StoreLocal:
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
