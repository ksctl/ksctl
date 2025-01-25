// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"bytes"
	"context"
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

	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/waiter"
	"golang.org/x/crypto/ssh"
)

type RemoteConnection interface {
	SSHExecute() error
	Flag(consts.KsctlUtilsConsts) RemoteConnection
	Script(ExecutionPipeline) RemoteConnection
	FastMode(bool) RemoteConnection
	Username(string)
	PrivateKey(string)
	GetOutput() []string
	IPv4(ip string) RemoteConnection
}

type SSH struct {
	UserName   string
	Privatekey string
	PublicIP   string
	Output     []string

	flag     consts.KsctlUtilsConsts
	script   ExecutionPipeline
	fastMode bool
	ctx      context.Context
	log      logger.Logger
}

func NewSSHExecutor(ctx context.Context, log logger.Logger, stateDocument *statefile.StorageDocument) *SSH {
	sshExecutor := &SSH{
		ctx: ctx,
		log: log,
	}
	sshExecutor.PrivateKey(stateDocument.K8sBootstrap.B.SSHInfo.PrivateKey)
	sshExecutor.Username(stateDocument.K8sBootstrap.B.SSHInfo.UserName)
	return sshExecutor
}

func (client *SSH) Username(s string) {
	client.UserName = s
}

func (client *SSH) PrivateKey(s string) {
	client.Privatekey = s
}

func (client *SSH) GetOutput() []string {
	out := make([]string, len(client.Output))
	copy(out, client.Output)
	client.Output = nil
	return out
}

func (client *SSH) IPv4(ip string) RemoteConnection {
	client.PublicIP = ip
	return client
}

func (client *SSH) Flag(execMethod consts.KsctlUtilsConsts) RemoteConnection {
	if execMethod == consts.UtilExecWithOutput || execMethod == consts.UtilExecWithoutOutput {
		client.flag = execMethod
		return client
	}
	return nil
}

func (client *SSH) Script(s ExecutionPipeline) RemoteConnection {
	client.script = s
	return client
}

func (client *SSH) FastMode(mode bool) RemoteConnection {
	client.fastMode = mode
	return client
}

func (client *SSH) ExecuteScript(conn *ssh.Client, script string) (stdout string, stderr string, err error) {

	if _, ok := config.IsContextPresent(client.ctx, consts.KsctlTestFlagKey); ok {
		return "stdout", "stderr", nil
	}

	expoBackoff := waiter.NewWaiter(
		5*time.Second,
		1,
		int(consts.CounterMaxNetworkSessionRetry),
	)

	_err := expoBackoff.Run(
		client.ctx,
		client.log,
		func() (err error) {
			stdout, stderr, err = func() (_stdout, _stderr string, _err error) {
				var _session *ssh.Session

				_session, _err = conn.NewSession()
				if _err != nil {
					return
				}

				defer func(_session *ssh.Session) {
					_ = _session.Close()
				}(_session)

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
			missingStatusErr := new(ssh.ExitMissingError)
			channelErr := new(ssh.OpenChannelError)

			if errors.As(err, &missingStatusErr) {
				client.log.Warn(client.ctx, "Retrying! Missing error code but exited", "Reason", err)
				return nil, false
			} else if errors.As(err, &channelErr) {
				client.log.Warn(client.ctx, "Retrying! Facing some channel open issues", "Reason", err)
				return nil, false
			} else {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrSSHExec,
					client.log.NewError(client.ctx, err.Error()),
				), true
			}
		},
		func() error {
			return nil
		},
		"Retrying, failed in client execution",
	)
	if _err != nil {
		err = _err
	}

	return
}

func (client *SSH) SSHExecute() error {

	privateKeyBytes := []byte(client.Privatekey)

	// create signer
	signer, err := signerFromPem(client.ctx, client.log, privateKeyBytes)
	if err != nil {
		return err
	}

	client.log.Debug(client.ctx, "SSH into", "sshAddr", fmt.Sprintf("%s@%s", client.UserName, client.PublicIP))

	c := &ssh.ClientConfig{
		User: client.UserName,
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
					recvFingerprint, err := returnServerPublicKeys(client.PublicIP, keyType)
					if err != nil {
						return ksctlErrors.WrapError(
							ksctlErrors.ErrSSHExec,
							client.log.NewError(
								client.ctx,
								"failed to fetch server public keys",
								err,
							),
						)
					}
					if recvFingerprint != gotFingerprint {
						return ksctlErrors.WrapError(
							ksctlErrors.ErrSSHExec,
							client.log.NewError(client.ctx, "mismatch of SSH fingerprint"),
						)
					}
					return nil
				}
				return ksctlErrors.WrapError(
					ksctlErrors.ErrSSHExec,
					client.log.NewError(client.ctx, "unsupported key type", "keyType", keyType),
				)
			})}

	if !client.fastMode {
		time.Sleep(consts.DurationSSHPause)
	}

	var conn *ssh.Client

	t0 := 5 * time.Second
	multT := 2
	maxRT := int(consts.CounterMaxRetryCount)
	if _, ok := config.IsContextPresent(client.ctx, consts.KsctlTestFlagKey); ok {
		t0 = time.Second
		multT = 1
		maxRT = 3
	}

	expoBackoff := waiter.NewWaiter(
		t0,
		multT,
		maxRT,
	)

	_err := expoBackoff.Run(
		client.ctx,
		client.log,
		func() (err error) {
			conn, err = ssh.Dial("tcp", client.PublicIP+":22", c)
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrSSHExec,
					client.log.NewError(client.ctx, "failed to get", "Reason", err))
			}
			return nil
		},
		func() bool {
			return true
		},
		nil,
		func() error {
			client.log.Note(client.ctx, "client tcp conn dial was successful")
			return nil
		},
		"Retrying, failed to establish client tcp conn dial",
	)
	if _err != nil {
		if _, ok := config.IsContextPresent(client.ctx, consts.KsctlTestFlagKey); ok {
			client.log.Note(client.ctx, "skipping the test for the client connection error", "Err", _err)
		} else {
			return _err
		}
	}

	client.log.Debug(client.ctx, "Printing", "bashScript", client.script)
	client.log.Print(client.ctx, "Exec Scripts")

	if _, ok := config.IsContextPresent(client.ctx, consts.KsctlTestFlagKey); !ok {
		defer func(conn *ssh.Client) {
			_ = conn.Close()
		}(conn)
	}

	scripts := client.script

	for !scripts.IsCompleted() {
		script := scripts.NextScript()

		client.log.Print(client.ctx, "Executing Sub-Script", "name", script.Name)
		client.log.Debug(client.ctx, "Script To Exec", script.ShellScript)
		success := false
		var scriptFailureReason error
		var stdout, stderr string
		var err error

		if script.CanRetry {
			retries := uint8(0)

			for retries < script.MaxRetries {
				stdout, stderr, err = client.ExecuteScript(conn, script.ShellScript)
				// adding some choas //
				if _, ok := config.IsContextPresent(client.ctx, consts.KsctlTestFlagKey); ok {
					if retries+1 < script.MaxRetries {
						err = client.log.NewError(client.ctx, "creating a fake choas error")
					}
				}
				/////////////////////
				if err != nil {
					client.log.Warn(client.ctx, "Failure in executing script", "retryCount", retries)
					scriptFailureReason = client.log.NewError(client.ctx, "Execute Failure", "stderr", stderr, "Reason", err)
					<-time.After(time.Duration(mrand.Intn(2)+1) * time.Second)
				} else {
					client.log.Debug(client.ctx, "client outputs", "stdout", stdout)
					success = true
					break
				}
				retries++
			}

		} else {
			stdout, stderr, err = client.ExecuteScript(conn, script.ShellScript)
			if err != nil {
				scriptFailureReason = client.log.NewError(client.ctx, "Failure in executing script", "Reason", err, "stderr", stderr)
			} else {
				success = true
				client.log.Debug(client.ctx, "client outputs", "stdout", stdout)
			}
		}

		if !success {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrSSHExec,
				scriptFailureReason)
		}
		if client.flag == consts.UtilExecWithOutput {
			client.Output = append(client.Output, stdout)
		}
	}

	client.log.Success(client.ctx, "Successful in executing the script")

	return nil
}

func signerFromPem(ctx context.Context, log logger.Logger, pemBytes []byte) (ssh.Signer, error) {

	// read pem block
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrSSHExec,
			log.NewError(ctx, "pem decode failed, no key found"),
		)
	}
	if x509.IsEncryptedPEMBlock(pemBlock) {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrSSHExec,
			log.NewError(ctx, "pem file is encrypted"),
		)
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrSSHExec,
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
