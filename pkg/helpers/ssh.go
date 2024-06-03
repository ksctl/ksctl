package helpers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"os/exec"
	"strings"
	"time"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	"golang.org/x/crypto/ssh"
)

type SSHPayload struct {
	UserName   string
	Privatekey string
	PublicIP   string
	Output     []string

	flag     consts.KsctlUtilsConsts
	script   types.ScriptCollection
	fastMode bool
	ctx      context.Context
	log      types.LoggerFactory
}

func NewSSHExecutor(ctx context.Context, log types.LoggerFactory, mainStateDocument *storageTypes.StorageDocument) *SSHPayload {
	sshExecutor := &SSHPayload{
		ctx: ctx,
		log: log,
	}
	sshExecutor.PrivateKey(mainStateDocument.K8sBootstrap.B.SSHInfo.PrivateKey)
	sshExecutor.Username(mainStateDocument.K8sBootstrap.B.SSHInfo.UserName)
	return sshExecutor
}

type SSHCollection interface {
	SSHExecute() error
	Flag(consts.KsctlUtilsConsts) SSHCollection
	Script(types.ScriptCollection) SSHCollection
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

func (sshPayload *SSHPayload) Script(s types.ScriptCollection) SSHCollection {
	sshPayload.script = s
	return sshPayload
}

func (sshPayload *SSHPayload) FastMode(mode bool) SSHCollection {
	sshPayload.fastMode = mode
	return sshPayload
}

func (sshExec *SSHPayload) ExecuteScript(conn *ssh.Client, script string) (stdout string, stderr string, err error) {

	if _, ok := IsContextPresent(sshExec.ctx, consts.KsctlTestFlagKey); ok {
		return "stdout", "stderr", nil
	}

	expoBackoff := NewBackOff(
		5*time.Second,
		1,
		int(consts.CounterMaxNetworkSessionRetry),
	)

	_err := expoBackoff.Run(
		sshExec.ctx,
		sshExec.log,
		func() (err error) {
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
			return err
		},
		func() bool {
			return true
		},
		func(err error) (errW error, escalateErr bool) {
			missingStatusErr := &ssh.ExitMissingError{}
			channelErr := &ssh.OpenChannelError{}

			if errors.As(err, &missingStatusErr) {
				sshExec.log.Warn(sshExec.ctx, "Retrying! Missing error code but exited", "Reason", err)
				return nil, false
			} else if errors.As(err, &channelErr) {
				sshExec.log.Warn(sshExec.ctx, "Retrying! Facing some channel open issues", "Reason", err)
				return nil, false
			} else {
				return err, true
			}
		},
		func() error {
			return nil
		},
		"Retrying, failed in ssh execution",
	)
	if _err != nil {
		err = _err
	}

	return
}

func (sshPayload *SSHPayload) SSHExecute() error {

	privateKeyBytes := []byte(sshPayload.Privatekey)

	// create signer
	signer, err := signerFromPem(sshPayload.ctx, sshPayload.log, privateKeyBytes)
	if err != nil {
		return err
	}

	sshPayload.log.Debug(sshPayload.ctx, "SSH into", "sshAddr", fmt.Sprintf("%s@%s", sshPayload.UserName, sshPayload.PublicIP))

	config := &ssh.ClientConfig{
		User: sshPayload.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: 5 * time.Minute,

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
						return ksctlErrors.ErrSSHExec.Wrap(
							sshPayload.log.NewError(
								sshPayload.ctx,
								"failed to fetch server public keys",
								err,
							),
						)
					}
					if recvFingerprint != gotFingerprint {
						return ksctlErrors.ErrSSHExec.Wrap(
							sshPayload.log.NewError(sshPayload.ctx, "mismatch of SSH fingerprint"),
						)
					}
					return nil
				}
				return ksctlErrors.ErrSSHExec.Wrap(
					sshPayload.log.NewError(sshPayload.ctx, "unsupported key type", "keyType", keyType),
				)
			})}

	if !sshPayload.fastMode {
		time.Sleep(consts.DurationSSHPause)
	}

	var conn *ssh.Client

	t0 := 5 * time.Second
	multT := 2
	maxRT := int(consts.CounterMaxRetryCount)
	if _, ok := IsContextPresent(sshPayload.ctx, consts.KsctlTestFlagKey); ok {
		t0 = time.Second
		multT = 1
		maxRT = 3
	}

	expoBackoff := NewBackOff(
		t0,
		multT,
		maxRT,
	)

	_err := expoBackoff.Run(
		sshPayload.ctx,
		sshPayload.log,
		func() (err error) {
			conn, err = ssh.Dial("tcp", sshPayload.PublicIP+":22", config)
			return err
		},
		func() bool {
			return true
		},
		nil,
		func() error {
			sshPayload.log.Note(sshPayload.ctx, "ssh tcp conn dial was successful")
			return nil
		},
		"Retrying, failed to establish ssh tcp conn dial",
	)
	if _err != nil {
		if _, ok := IsContextPresent(sshPayload.ctx, consts.KsctlTestFlagKey); ok {
			sshPayload.log.Note(sshPayload.ctx, "skipping the test for the ssh connection error", "Err", _err)
		} else {
			return _err
		}
	}

	sshPayload.log.Debug(sshPayload.ctx, "Printing", "bashScript", sshPayload.script)
	sshPayload.log.Print(sshPayload.ctx, "Exec Scripts")

	if _, ok := IsContextPresent(sshPayload.ctx, consts.KsctlTestFlagKey); !ok {
		defer conn.Close()
	}

	scripts := sshPayload.script

	for !scripts.IsCompleted() {
		script := scripts.NextScript()

		sshPayload.log.Print(sshPayload.ctx, "Executing Sub-Script", "name", script.Name)
		sshPayload.log.Debug(sshPayload.ctx, "Script To Exec", script.ShellScript)
		success := false
		var scriptFailureReason error
		var stdout, stderr string
		var err error

		if script.CanRetry {
			retries := uint8(0)

			for retries < script.MaxRetries {
				stdout, stderr, err = sshPayload.ExecuteScript(conn, script.ShellScript)
				// adding some choas //
				if _, ok := IsContextPresent(sshPayload.ctx, consts.KsctlTestFlagKey); ok {
					if retries+1 < script.MaxRetries {
						err = sshPayload.log.NewError(sshPayload.ctx, "creating a fake choas error")
					}
				}
				/////////////////////
				if err != nil {
					sshPayload.log.Warn(sshPayload.ctx, "Failure in executing script", "retryCount", retries)
					scriptFailureReason = sshPayload.log.NewError(sshPayload.ctx, "Execute Failure", "stderr", stderr, "Reason", err)
					<-time.After(time.Duration(mrand.Intn(2)+1) * time.Second)
				} else {
					sshPayload.log.Debug(sshPayload.ctx, "ssh outputs", "stdout", stdout)
					success = true
					break
				}
				retries++
			}

		} else {
			stdout, stderr, err = sshPayload.ExecuteScript(conn, script.ShellScript)
			if err != nil {
				scriptFailureReason = sshPayload.log.NewError(sshPayload.ctx, "Failure in executing script", "Reason", err, "stderr", stderr)
			} else {
				success = true
				sshPayload.log.Debug(sshPayload.ctx, "ssh outputs", "stdout", stdout)
			}
		}

		if !success {
			return ksctlErrors.ErrSSHExec.Wrap(scriptFailureReason)
		}
		if sshPayload.flag == consts.UtilExecWithOutput {
			sshPayload.Output = append(sshPayload.Output, stdout)
		}
	}

	sshPayload.log.Success(sshPayload.ctx, "Successful in executing the script")

	return nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(ctx context.Context, log types.LoggerFactory, bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to generate key", "Reason", err),
		)
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to validate key", "Reason", err),
		)
	}

	log.Print(ctx, "Private Key helper-gen")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(log types.LoggerFactory, privateKey *rsa.PrivateKey) []byte {
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
func generatePublicKey(ctx context.Context, log types.LoggerFactory, privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "failed to create public key for given private key", "Reason", err),
		)
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Print(ctx, "Public key helper-gen")
	return pubKeyBytes, nil
}

func CreateSSHKeyPair(ctx context.Context, log types.LoggerFactory, state *storageTypes.StorageDocument) error {

	bitSize := 4096

	privateKey, err := generatePrivateKey(ctx, log, bitSize)
	if err != nil {
		return err
	}

	publicKeyBytes, err := generatePublicKey(ctx, log, &privateKey.PublicKey)
	if err != nil {
		return err
	}

	privateKeyBytes := encodePrivateKeyToPEM(log, privateKey)

	log.Debug(ctx, "Printing", "ssh pub key", string(publicKeyBytes))
	log.Debug(ctx, "Printing", "ssh private key", string(privateKeyBytes))

	state.SSHKeyPair.PrivateKey = string(privateKeyBytes)
	state.SSHKeyPair.PublicKey = string(publicKeyBytes)

	return nil
}

func signerFromPem(ctx context.Context, log types.LoggerFactory, pemBytes []byte) (ssh.Signer, error) {

	// read pem block
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "pem decode failed, no key found"),
		)
	}
	if x509.IsEncryptedPEMBlock(pemBlock) {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "pem file is encrypted"),
		)
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, ksctlErrors.ErrSSHExec.Wrap(
			log.NewError(ctx, "parsing plain private key failed", "Reason", err),
		)
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
