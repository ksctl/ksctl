package helpers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"golang.org/x/crypto/ssh"
)

type SSHPayload struct {
	UserName   string
	Privatekey string
	PublicIP   string
	Output     []string

	flag     consts.KsctlUtilsConsts
	script   resources.ScriptCollection
	fastMode bool
}

func NewSSHExecutor(mainStateDocument *types.StorageDocument) SSHCollection {
	var sshExecutor SSHCollection = &SSHPayload{}
	sshExecutor.PrivateKey(mainStateDocument.K8sBootstrap.B.SSHInfo.PrivateKey)
	sshExecutor.Username(mainStateDocument.K8sBootstrap.B.SSHInfo.UserName)
	return sshExecutor
}

type SSHCollection interface {
	SSHExecute(resources.LoggerFactory) error
	Flag(consts.KsctlUtilsConsts) SSHCollection
	Script(resources.ScriptCollection) SSHCollection
	FastMode(bool) SSHCollection
	Username(string)
	PrivateKey(string)
	GetOutput() []string
	IPv4(ip string) SSHCollection
}

func (sshPayload *SSHPayload) Username(s string) {
	sshPayload.UserName = s
}

func (sshPayload *SSHPayload) PrivateKey(s string) {
	sshPayload.Privatekey = s
}

func (sshPayload *SSHPayload) GetOutput() []string {
	out := make([]string, len(sshPayload.Output))
	copy(out, sshPayload.Output)
	sshPayload.Output = nil
	return out
}

func (sshPayload *SSHPayload) IPv4(ip string) SSHCollection {
	sshPayload.PublicIP = ip
	return sshPayload
}

func (sshPayload *SSHPayload) Flag(execMethod consts.KsctlUtilsConsts) SSHCollection {
	if execMethod == consts.UtilExecWithOutput || execMethod == consts.UtilExecWithoutOutput {
		sshPayload.flag = execMethod
		return sshPayload
	}
	return nil
}

func (sshPayload *SSHPayload) Script(s resources.ScriptCollection) SSHCollection {
	sshPayload.script = s
	return sshPayload
}

func (sshPayload *SSHPayload) FastMode(mode bool) SSHCollection {
	sshPayload.fastMode = mode
	return sshPayload
}

func ExecuteScript(log resources.LoggerFactory, conn *ssh.Client, script string) (stdout string, stderr string, err error) {

	netRetry := 0
	for netRetry < int(consts.CounterMaxNetworkSessionRetry) {

		stdout, stderr, err = func() (_stdout, _stderr string, _err error) {
			var _session *ssh.Session

			_session, _err = conn.NewSession()
			if _err != nil {
				return
			}

			defer _session.Close()

			_bout := new(strings.Builder)
			_berr := new(strings.Builder)
			_session.Stdout = _bout
			_session.Stderr = _berr

			defer func() {
				_stdout, _stderr = _bout.String(), _berr.String()
			}()

			_err = _session.Run(script)
			return
		}()

		if err == nil {
			return
		}

		if err != nil {
			errV := ssh.ExitMissingError{}
			scriptErrV := ssh.ExitError{}

			fmt.Printf("Error checks err: %#v, errV: %#v, scriptErrV: %#v\n", err, errV, scriptErrV)
			fmt.Printf("Error Erro() err: %#v, errV: %#v, scriptErrV: %#v\n", err.Error(), errV.Error(), scriptErrV.Error())

			switch err.Error() {
			case errV.Error():
				log.Warn("Missing error code but exited. Reason can be session comm failure. Retrying!")
				netRetry++
			default:
				// case scriptErrV.Error():
				return // if any error which is not same as ExitMissingError
			}
		}
	}

	if netRetry == int(consts.CounterMaxNetworkSessionRetry) {
		err = log.NewError("Retry failed with network problems: %v", err)
	}
	return
}

func (sshPayload *SSHPayload) SSHExecute(log resources.LoggerFactory) error {

	privateKeyBytes := []byte(sshPayload.Privatekey)

	// create signer
	signer, err := signerFromPem(privateKeyBytes)
	if err != nil {
		return err
	}
	log.Debug("SSH into", "sshAddr", fmt.Sprintf("%s@%s", sshPayload.UserName, sshPayload.PublicIP))

	if fake := os.Getenv(string(consts.KsctlFakeFlag)); len(fake) != 0 {
		log.Debug("Exec Scripts for fake flag")
		sshPayload.Output = []string{}

		if sshPayload.flag == consts.UtilExecWithOutput {
			sshPayload.Output = []string{"random fake"}
		}
		return nil
	}

	config := &ssh.ClientConfig{
		User: sshPayload.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: time.Duration(5 * time.Minute),

		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSASHA256,
			ssh.KeyAlgoED25519,
		},
		HostKeyCallback: ssh.HostKeyCallback(
			func(hostname string, remote net.Addr, remoteSvrHostKey ssh.PublicKey) error {
				gotFingerprint := ssh.FingerprintSHA256(remoteSvrHostKey)
				keyType := remoteSvrHostKey.Type()
				if keyType == ssh.KeyAlgoRSA || keyType == ssh.KeyAlgoED25519 {
					recvFingerprint, err := returnServerPublicKeys(sshPayload.PublicIP, keyType)
					if err != nil {
						return err
					}
					if recvFingerprint != gotFingerprint {
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
			log.Warn("RETRYING", "reason", err)
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

	scripts := sshPayload.script

	for !scripts.IsCompleted() {
		script := scripts.NextScript()

		log.Print("Executing Sub-Script", "name", script.Name)
		success := false
		var scriptFailureReason error
		var stdout, stderr string
		var err error

		if script.CanRetry {
			retries := uint8(0)

			for retries < script.MaxRetries {
				stdout, stderr, err = ExecuteScript(log, conn, script.ShellScript)
				if err != nil {
					log.Warn("Failure in executing script", "retryCount", retries)
					scriptFailureReason = log.NewError("Execute Failure", "stderr", stderr)
				} else {
					log.Debug("ssh outputs", "stdout", stdout)
					success = true
					break
				}
				retries++
			}

		} else {
			stdout, stderr, err = ExecuteScript(log, conn, script.ShellScript)
			if err != nil {
				log.Error("Failure in executing script", "Reason", err)
				scriptFailureReason = log.NewError("Execute Failure", "stderr", stderr)
			} else {
				success = true
				log.Debug("ssh outputs", "stdout", stdout)
			}
		}

		if !success {
			return scriptFailureReason
		}
		if sshPayload.flag == consts.UtilExecWithOutput {
			sshPayload.Output = append(sshPayload.Output, stdout)
		}
	}

	log.Success("Successful in executing the script")

	return nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(log resources.LoggerFactory, bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Print("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(log resources.LoggerFactory, privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(log resources.LoggerFactory, privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Print("Public key generated")
	return pubKeyBytes, nil
}
func CreateSSHKeyPair(log resources.LoggerFactory, state *types.StorageDocument) error {

	bitSize := 4096

	privateKey, err := generatePrivateKey(log, bitSize)
	if err != nil {
		return err
	}

	publicKeyBytes, err := generatePublicKey(log, &privateKey.PublicKey)
	if err != nil {
		return err
	}

	privateKeyBytes := encodePrivateKeyToPEM(log, privateKey)

	log.Debug("Printing", "ssh pub key", string(publicKeyBytes))
	log.Debug("Printing", "ssh private key", string(privateKeyBytes))

	state.SSHKeyPair.PrivateKey = string(privateKeyBytes)
	state.SSHKeyPair.PublicKey = string(publicKeyBytes)

	return nil
}

func signerFromPem(pemBytes []byte) (ssh.Signer, error) {

	// read pem block
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, errors.New("pem decode failed, no key found")
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

// returnServerPublicKeys it uses the ssh-keygen and ssh-keyscan as OS deps
// it uses this command -> ssh-keyscan -t rsa <remote_ssh_server_public_ipv4> | ssh-keygen -lf -
func returnServerPublicKeys(publicIP string, keyType string) (string, error) {
	var c1, c2 *exec.Cmd

	switch keyType {
	case ssh.KeyAlgoRSA:
		c1 = exec.Command("ssh-keyscan", "-t", "rsa", publicIP)
	case ssh.KeyAlgoED25519:
		c1 = exec.Command("ssh-keyscan", "-t", "ed25519", publicIP)
	}

	c2 = exec.Command("ssh-keygen", "-lf", "-")

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
