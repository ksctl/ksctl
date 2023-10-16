package universal

import (
	"fmt"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

type InstallType string

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)

type Application struct {
	Name        string
	Url         string
	Namespace   string
	Version     string
	Metadata    string
	Maintainer  string
	PostInstall string
	HelmConfig  []WorkLoad
	InstallType
}

func (a Application) String() string {
	return fmt.Sprintf(`
Name: %s
Metadata: %s
Namespace: %s
Version: %s
Maintainer: %s
Source: %s
InstallType: %s
PostInstall: %s
`, a.Name, a.Metadata, a.Namespace, a.Version, a.Maintainer, a.Url, a.InstallType, a.PostInstall)
}

var (
	apps map[string]func() Application
)

func initApps() {
	apps = map[string]func() Application{
		"argo-rollouts":    argoRolloutsData,
		"argocd":           argocdData,
		"istio":            istioData,
		"cilium":           ciliumData,
		"prometheus-stack": prometheusStackData,
	}
}

func GetApps(storage resources.StorageFactory, name string) (Application, error) {
	if apps == nil {
		return Application{}, fmt.Errorf("[kubernetes] app variable not initalized")
	}

	val, present := apps[name]

	if !present {
		return Application{}, fmt.Errorf("[kubernetes] app not found %s", name)
	}
	return val(), nil
}

func (this *Kubernetes) InstallCNI(app string) error {

	if err := installApplication(this, app); err != nil {
		return err
	}

	this.StorageDriver.Logger().Success("[kubernetes] Installed CNI plugin")
	return nil
}

func (this *Kubernetes) InstallApplications(apps []string) error {

	for _, app := range apps {
		if err := installApplication(this, app); err != nil {
			return err
		}
	}

	this.StorageDriver.Logger().Success("[kubernetes] Installed Applications")
	return nil
}

func installApplication(client *Kubernetes, app string) error {

	appStruct, err := GetApps(client.StorageDriver, app)
	if err != nil {
		return err
	}

	switch appStruct.InstallType {

	case InstallHelm:
		if err := installHelm(client, appStruct); err != nil {
			return err
		}

	case InstallKubectl:
		if err := installKubectl(client, appStruct); err != nil {
			return err
		}

	}

	client.StorageDriver.Logger().Success("[kubernetes] Installed Resource\n", appStruct.String())
	return nil
}
