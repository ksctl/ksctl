package kubernetes

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/resources"
)

type InstallType string

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)

type Application struct {
	Name          string
	Url           string
	Namespace     string
	Version       string
	Metadata      string
	Maintainer    string
	PostInstall   string
	HelmConfig    []HelmOptions
	KubectlConfig KubectlOptions
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
		"ksctl-storage":    storageImportData,
	}
}

func GetApps(storage resources.StorageFactory, name string) (Application, error) {
	if apps == nil {
		return Application{}, log.NewError("app variable not initalized")
	}

	val, present := apps[name]

	if !present {
		return Application{}, log.NewError("app not found %s", name)
	}
	return val(), nil
}

func (this *Kubernetes) InstallCNI(app string) error {

	if err := installApplication(this, app); err != nil {
		return log.NewError(err.Error())
	}

	log.Success("Installed CNI plugin")
	return nil
}

// InstallApplications Important the sequence of the apps in the list are important
// it executes from left to right one at a time
// if it fails at any point of time it stop further installations
func (this *Kubernetes) InstallApplications(apps []string) error {

	for _, app := range apps {
		if err := installApplication(this, app); err != nil {
			return log.NewError(err.Error())
		}
	}

	log.Success("Installed Applications")
	return nil
}

func installApplication(client *Kubernetes, app string) error {

	appStruct, err := GetApps(client.StorageDriver, app)
	if err != nil {
		return log.NewError(err.Error())
	}

	switch appStruct.InstallType {

	case InstallHelm:
		if err := installHelm(client, appStruct); err != nil {
			return log.NewError(err.Error())
		}

	case InstallKubectl:
		if err := installKubectl(client, appStruct); err != nil {
			return log.NewError(err.Error())
		}

	}

	log.Success("Installed Resource", "metadata", appStruct.String())
	return nil
}
