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

package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

// JoinWorkerplane implements storage.DistroFactory.
func (k3s *K3s) JoinWorkerplane(no int, _ types.StorageFactory) error {
	k3s.mu.Lock()
	idx := no
	sshExecutor := helpers.NewSSHExecutor(k3sCtx, log, mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	k3s.mu.Unlock()

	log.Note(k3sCtx, "configuring Workerplane", "number", strconv.Itoa(idx))

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptWP(
			mainStateDocument.K8sBootstrap.K3s.K3sVersion,
			mainStateDocument.K8sBootstrap.B.PrivateIPs.LoadBalancer,
			mainStateDocument.K8sBootstrap.K3s.K3sToken,
		)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.WorkerPlanes[idx]).
		FastMode(true).SSHExecute()
	if err != nil {
		return err
	}

	log.Success(k3sCtx, "configured WorkerPlane", "number", strconv.Itoa(idx))

	return nil
}

func scriptWP(ver string, privateIPlb, token string) types.ScriptCollection {

	collection := helpers.NewScriptCollection()

	collection.Append(types.Script{
		Name:           "Join the workerplane-[0..M]",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
cat <<EOF > worker-setup.sh
#!/bin/bash
/bin/bash /usr/local/bin/k3s-agent-uninstall.sh || echo "already deleted"
export K3S_DEBUG=true
curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL="%s" sh -s - agent --token %s --server https://%s:6443
EOF

sudo chmod +x worker-setup.sh
sudo ./worker-setup.sh &>> ksctl.log
`, ver, token, privateIPlb),
	})

	return collection
}
