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

	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"golang.org/x/crypto/ssh"
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
	SSHExecute(resources.StorageFactory, resources.LoggerFactory, consts.KsctlCloud) error
	Flag(consts.KsctlUtilsConsts) SSHCollection
	Script(string) SSHCollection
	FastMode(bool) SSHCollection
	Username(string)
	LocPrivateKey(string)
	GetOutput() string
	IPv4(ip string) SSHCollection
}

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

func (sshPayload *SSHPayload) SSHExecute(storage resources.StorageFactory, log resources.LoggerFactory, provider consts.KsctlCloud) error {

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
			func(hostname string, remote net.Addr, remoteSvrHostKey ssh.PublicKey) error {
				gotFingerprint := ssh.FingerprintSHA256(remoteSvrHostKey)
				keyType := remoteSvrHostKey.Type()
				if keyType == ssh.KeyAlgoRSA {
					recvFingerprint, err := returnServerPublicKeys(sshPayload.PublicIP, provider)
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

	if provider == consts.CloudAws {
		config = &ssh.ClientConfig{
			User: "ubuntu",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},

			HostKeyAlgorithms: []string{
				ssh.KeyAlgoECDSA256,
				ssh.KeyAlgoED25519,
				// ssh.KeyAlgoRSA,
			},
			HostKeyCallback: ssh.HostKeyCallback(
				func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					actualFingerprint := ssh.FingerprintSHA256(key)
					keyType := key.Type()
					if keyType == ssh.KeyAlgoRSA || keyType == ssh.KeyAlgoDSA || keyType == ssh.KeyAlgoECDSA256 || keyType == ssh.KeyAlgoED25519 {
						expectedFingerprint, err := returnServerPublicKeys(sshPayload.PublicIP, provider)
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
	}
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
func CreateSSHKeyPair(storage resources.StorageFactory, log resources.LoggerFactory, provider consts.KsctlCloud, clusterDir string) (string, error) {
	savePrivateFileTo := GetPath(consts.UtilOtherPath, provider, consts.ClusterTypeHa, clusterDir, "keypair")
	savePublicFileTo := GetPath(consts.UtilOtherPath, provider, consts.ClusterTypeHa, clusterDir, "keypair.pub")

	bitSize := 4096

	privateKey, err := generatePrivateKey(log, bitSize)
	if err != nil {
		return "", err
	}

	publicKeyBytes, err := generatePublicKey(log, &privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	privateKeyBytes := encodePrivateKeyToPEM(log, privateKey)

	log.Debug("Printing", "ssh pub key", string(publicKeyBytes))
	log.Debug("Printing", "ssh private key", string(privateKeyBytes))

	if err := storage.Path(savePrivateFileTo).Permission(0400).Save(privateKeyBytes); err != nil {
		return "", err
	}

	if err := storage.Path(savePublicFileTo).Permission(0600).Save(publicKeyBytes); err != nil {
		return "", err
	}

	return string(publicKeyBytes), nil
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

// returnServerPublicKeys it uses the ssh-keygen and ssh-keyscan as OS deps
// it uses this command -> ssh-keyscan -t rsa <remote_ssh_server_public_ipv4> | ssh-keygen -lf -
func returnServerPublicKeys(publicIP string, provider consts.KsctlCloud) (string, error) {
	var c1 *exec.Cmd
	var c2 *exec.Cmd

	c1 = exec.Command("ssh-keyscan", "-t", "rsa", publicIP) // WARN: it requires the os to have these dependencies
	c2 = exec.Command("ssh-keygen", "-lf", "-")             // WARN: it requires the os to have these dependencies

	if provider == consts.CloudAws {
		c1 = exec.Command("ssh-keyscan", "-t", "ecdsa", publicIP)
		c2 = exec.Command("ssh-keygen", "-lf", "-")
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
