package kubernetes

var (
	apps map[string]func(string) Application
)

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)
