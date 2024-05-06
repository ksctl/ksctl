package kubernetes

import (
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

type InstallType string

const (
	InstallKubectl = InstallType("kubectl")
	InstallHelm    = InstallType("helm")
)

type Application struct {
	Name          string
	Url           string
	Version       string
	Metadata      string
	Maintainer    string
	PostInstall   string
	HelmConfig    []HelmOptions
	KubectlConfig KubectlOptions
	InstallType
}

var (
	apps map[string]func(string) Application
)

func initApps() {
	apps = map[string]func(string) Application{
		"argo-rollouts":     argoRolloutsData,
		"argocd":            argocdData,
		"istio":             istioData,
		"cilium":            ciliumData,
		"prometheus-stack":  prometheusStackData,
		"ksctl-storage":     storageImportData,
		"ksctl-application": applicationStackData,
		"flannel":           flannelData,
	}
}

func GetApps(name string, ver string) (Application, error) {
	if apps == nil {
		return Application{}, log.NewError("app variable not initalized")
	}

	val, present := apps[name]

	if !present {
		return Application{}, log.NewError("app not found %s", name)
	}
	return val(ver), nil
}

type EnumApplication string

const (
	Cni EnumApplication = "cni"
	App EnumApplication = "app"
)

// NOTE: updatable means the app is present and we can upgrade or degrade the version as per users wish

func PresentOrNot(app types.Application, typeOfApp EnumApplication, state *types.StorageDocument) (idx int, isPresent bool) {
	idx = -1

	installedApps := state.Addons

	switch typeOfApp {
	case Cni:
		if app.Name == installedApps.Cni.Name {
			isPresent = true
			return
		}
	case App:
		for _idx, _app := range installedApps.Apps {
			if _app.Name == app.Name {
				idx = _idx
				isPresent = true
				return
			}
		}
	}

	return
}

func (k *Kubernetes) InstallCNI(cni types.Application, state *types.StorageDocument, op consts.KsctlOperation) error {

	switch op {
	case consts.OperationCreate:
		_, ok := PresentOrNot(cni, Cni, state)
		if ok {
			if cni.Version == state.Addons.Cni.Version {
				log.Success("Already Installed cni", "name", cni.Name, "version", cni.Version)
				return nil
			} else {
				if k.InCluster {
					return log.NewError("We cannot install CNI due to Operation inside the cluster", "name", cni.Name, "version", cni.Version)
				} else {
					log.Box("Current Impl. doesn't support k", `
Upgrade of CNI is not Possible as of now!
Reason: if the cni is uninstalled it will lead to all pod in Pending mode
thus we can't install cni without the help of state.
So what we can do is Delete it and then

solution is instead of performing k operation inside the cluster which will become hostile
will perform k only from outside like the ksctl core for the cli or UI
so what we can do is we can tell ksctl core to fetch latest state and then we can perform operations

another nice thing would be to reconcile every 2 or 5 minutes from the kubernetes cluster Export()
	(Only k problem will occur for local based system)
advisiable to use external storage solution
`)
				}

				// Step 1: delete the current install
				// Step 2: install with the new Version
				return nil // saftey return to avoid version conflicts, k feature is yet to be developed
			}
		}

		if err := installApplication(k, cni); err != nil {
			return log.NewError("Cni install failed", "name", cni, "errorMsg", err)
		}
		state.Addons.Cni.Name = cni.Name
		state.Addons.Cni.Version = cni.Version

		if err := k.StorageDriver.Write(state); err != nil {
			return err
		}

		log.Success("Installed Cni")

	case consts.OperationDelete:

		_, ok := PresentOrNot(cni, App, state)
		if !ok {
			log.Success("Cni is not present", "name", cni.Name, "version", cni.Version)
			return nil
		}

		if err := deleteApplication(k, cni); err != nil {
			return log.NewError("Cni uninstall failed", "name", cni, "errorMsg", err)
		}

		state.Addons.Cni.Name = ""
		state.Addons.Cni.Version = ""

		if err := k.StorageDriver.Write(state); err != nil {
			return err
		}

		log.Success("Uninstalled Cni")
	}

	return nil
}

// Applications Important the sequence of the apps in the list are important
// it executes from left to right one at a time
// if it fails at any point of time it stop further installations
func (k *Kubernetes) Applications(apps []types.Application, state *types.StorageDocument, op consts.KsctlOperation) error {

	switch op {
	case consts.OperationCreate:
		for idx, app := range apps {
			_idx, ok := PresentOrNot(app, App, state)
			isUpdate := false
			prevVersion := ""

			if ok {
				if app.Version == state.Addons.Apps[_idx].Version {
					log.Success("Already Installed app", "name", app.Name, "version", app.Version)
					continue
				} else {
					// Delete the App
					isUpdate = true
					prevVersion = state.Addons.Apps[_idx].Version
					if err := deleteApplication(k, state.Addons.Apps[_idx]); err != nil {
						return log.NewError("Update of the App failed Step Uninstall",
							"app", app.Name,
							"FromVer", prevVersion,
							"ToVer", app.Version,
							"errorMsg", err)
					}

					// Install the App in the same flow (allowing to flow down)
				}
			}

			if err := installApplication(k, app); err != nil {
				if isUpdate {
					return log.NewError("Update of the App failed Step Install",
						"app", app.Name,
						"FromVer", prevVersion,
						"ToVer", app.Version,
						"errorMsg", err)
				}
				return log.NewError("App install failed", "app", app, "errorMsg", err)
			}
			if isUpdate {
				state.Addons.Apps[_idx].Version = app.Version
			} else {
				state.Addons.Apps = append(state.Addons.Apps, types.Application{
					Name:    app.Name,
					Version: app.Version,
				})
			}
			if err := k.StorageDriver.Write(state); err != nil {
				return err
			}

			if isUpdate {
				log.Success("Updated the App",
					"app", app.Name,
					"FromVer", prevVersion,
					"ToVer", app.Version)
			}
			log.Success("Installed Application", "Success", idx+1, "Total", len(apps))
		}

	case consts.OperationDelete:
		for idx, app := range apps {

			_idx, ok := PresentOrNot(app, App, state)
			if !ok {
				log.Success("App is not present", "name", app.Name, "version", app.Version)
				continue
			}

			if err := deleteApplication(k, app); err != nil {
				return log.NewError("App uninstall failed", "app", app, "errorMsg", err)
			}

			_cpyApp := utilities.DeepCopySlice[types.Application](state.Addons.Apps)
			for _i, _app := range state.Addons.Apps {
				if _i != _idx {
					_cpyApp = append(_cpyApp, _app)
				}
			}
			state.Addons.Apps = _cpyApp
			if err := k.StorageDriver.Write(state); err != nil {
				return err
			}

			log.Success("Uninstalled Application", "Success", idx+1, "Total", len(apps))
		}
	}

	return nil
}

func installApplication(client *Kubernetes, app types.Application) error {

	if err := helpers.IsValidVersion(app.Version); err != nil {
		return log.NewError(err.Error())
	}

	appStruct, err := GetApps(app.Name, app.Version)
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

	log.Box("App Details", appStruct.Metadata+"\n"+appStruct.PostInstall)

	log.Success("Installed Resource")
	return nil
}

func deleteApplication(client *Kubernetes, app types.Application) error {

	if err := helpers.IsValidVersion(app.Version); err != nil {
		return log.NewError(err.Error())
	}
	appStruct, err := GetApps(app.Name, app.Version)
	if err != nil {
		return log.NewError(err.Error())
	}

	switch appStruct.InstallType {

	case InstallHelm:
		if err := deleteHelm(client, appStruct); err != nil {
			return log.NewError(err.Error())
		}

	case InstallKubectl:
		if err := deleteKubectl(client, appStruct); err != nil {
			return log.NewError(err.Error())
		}

	}

	log.Success("Uninstalled Resource")
	return nil
}
