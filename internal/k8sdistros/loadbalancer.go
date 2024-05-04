package k8sdistros

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"

	"github.com/ksctl/ksctl/pkg/resources"
)

func (p *PreBootstrap) ConfigureLoadbalancer(_ resources.StorageFactory) error {
	log.Print("configuring Loadbalancer")
	p.mu.Lock()
	sshExecutor := helpers.NewSSHExecutor(mainStateDocument) //making sure that a new obj gets initialized for a every run thus eleminating possible problems with concurrency
	p.mu.Unlock()

	controlPlaneIPs := utilities.DeepCopySlice[string](mainStateDocument.K8sBootstrap.B.PublicIPs.ControlPlanes)

	err := sshExecutor.Flag(consts.UtilExecWithoutOutput).Script(
		scriptConfigureLoadbalancer(controlPlaneIPs)).
		IPv4(mainStateDocument.K8sBootstrap.B.PublicIPs.LoadBalancer).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}

	log.Success("configured LoadBalancer")
	return nil
}

func scriptConfigureLoadbalancer(controlPlaneIPs []string) resources.ScriptCollection {
	collection := helpers.NewScriptCollection()

	collection.Append(resources.Script{
		Name:       "Install haproxy",
		CanRetry:   true,
		MaxRetries: 9,
		ShellScript: `
sudo apt update -y
sudo apt install haproxy -y
`,
		ScriptExecutor: consts.LinuxBash,
	})

	collection.Append(resources.Script{
		Name:           "enable and start systemd service for haproxy",
		CanRetry:       true,
		MaxRetries:     3,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl start haproxy
sudo systemctl enable haproxy
`,
	})

	serverScript := ""
	for index, controlPlaneIP := range controlPlaneIPs {
		serverScript += fmt.Sprintf(`  server k3sserver-%d %s:%d check
`, index+1, controlPlaneIP, 6443)
	}

	nodePortScript := ""
	for index, controlPlaneIP := range controlPlaneIPs {
		nodePortScript += fmt.Sprintf(`  server k3sserver-%d %s
`, index+1, controlPlaneIP)
	}

	collection.Append(resources.Script{
		Name:           "create haproxy configuration",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: fmt.Sprintf(`
cat <<EOF > haproxy.cfg
frontend kubernetes-frontend
  bind *:6443
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend


frontend kubernetes-nodeport
  bind *:30000-35000
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend-nodeport

backend kubernetes-backend
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
%s

backend kubernetes-backend-nodeport
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
%s
EOF

sudo mv haproxy.cfg /etc/haproxy/haproxy.cfg
`, serverScript, nodePortScript),
	})

	collection.Append(resources.Script{
		Name:           "create haproxy configuration",
		CanRetry:       false,
		ScriptExecutor: consts.LinuxBash,
		ShellScript: `
sudo systemctl restart haproxy
`,
	})

	return collection
}
