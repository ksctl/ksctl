//////////////// THIS FILE IS NOT READY ////////////////

package utils

import (
	"bytes"
	"crypto/x509"
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

	"github.com/kubesimplify/ksctl/api/provider/logger"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

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

type SSHPayload struct {
	UserName       string `json:"user_name"`
	PathPrivateKey string `json:"path_private_key"`
	PublicIP       string `json:"public_ip"`
	Output         string `json:"output"`
}

// GetUserName returns current active username
func GetUserName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}
func IsValidName(clusterName string) bool {
	matched, _ := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, clusterName)

	return matched
}

// TODO: INcomplete
func CreateSSHKeyPair(provider, clusterDir string) (string, error) {

	pathTillFolder := "" // TODO: here
	// pathTillFolder := getPaths(provider, "ha", clusterDir)

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", "keypair")
	cmd.Dir = pathTillFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fmt.Println(string(out))

	keyPairToUpload := "" // TODO: here
	// keyPairToUpload := GetPath(OTHER_PATH, provider, "ha", clusterDir, "keypair.pub")
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

func UserInputCredentials(logging logger.Logger) (string, error) {

	fmt.Print("    Enter Secret-> ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
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
