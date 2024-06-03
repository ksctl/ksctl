package kubernetes

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

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
	Maintainer    string
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
		"ksctl-application": applicationStackData,
		"flannel":           flannelData,
	}
}

func GetApps(name string, ver string) (Application, error) {
	if apps == nil {
		return Application{}, log.NewError(kubernetesCtx, "app variable not initalized")
	}

	val, present := apps[name]

	if !present {
		return Application{}, log.NewError(kubernetesCtx, "app not found", "name", name)
	}
	return val(ver), nil
}

type EnumApplication string

const (
	Cni EnumApplication = "cni"
	App EnumApplication = "app"
)

// NOTE: updatable means the app is present and we can upgrade or degrade the version as per users wish

func PresentOrNot(app storageTypes.Application, typeOfApp EnumApplication, state *storageTypes.StorageDocument) (idx int, isPresent bool) {
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

func (k *Kubernetes) InstallCNI(cni storageTypes.Application, state *storageTypes.StorageDocument, op consts.KsctlOperation) error {

	switch op {
	case consts.OperationCreate:
		_, ok := PresentOrNot(cni, Cni, state)
		if ok {
			if cni.Version == state.Addons.Cni.Version {
				log.Success(kubernetesCtx, "Already Installed cni", "name", cni.Name, "version", cni.Version)
				return nil
			} else {
				if k.inCluster {
					return log.NewError(kubernetesCtx, "We cannot install CNI due to Operation inside the cluster", "name", cni.Name, "version", cni.Version)
				} else {
					log.Box(kubernetesCtx, "Current Impl. doesn't support cni upgrade", `
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
			return log.NewError(kubernetesCtx, "Cni install failed", "name", cni, "Reason", err)
		}
		state.Addons.Cni.Name = cni.Name
		state.Addons.Cni.Version = cni.Version

		if err := k.storageDriver.Write(state); err != nil {
			return err
		}

		log.Success(kubernetesCtx, "Installed Cni", "name", cni.Name, "version", cni.Version)

	case consts.OperationDelete:

		_, ok := PresentOrNot(cni, App, state)
		if !ok {
			log.Success(kubernetesCtx, "Cni is not present", "name", cni.Name, "version", cni.Version)
			return nil
		}

		if err := deleteApplication(k, cni); err != nil {
			return log.NewError(kubernetesCtx, "Cni uninstall failed", "name", cni, "Reason", err)
		}

		state.Addons.Cni.Name = ""
		state.Addons.Cni.Version = ""

		if err := k.storageDriver.Write(state); err != nil {
			return err
		}

		log.Success(kubernetesCtx, "Uninstalled Cni", "name", cni.Name, "version", cni.Version)
	}

	return nil
}

// Applications Important the sequence of the apps in the list are important
// it executes from left to right one at a time
// if it fails at any point of time it stop further installations
func (k *Kubernetes) Applications(apps []storageTypes.Application, state *storageTypes.StorageDocument, op consts.KsctlOperation) error {

	switch op {
	case consts.OperationCreate:
		for idx, app := range apps {
			_idx, ok := PresentOrNot(app, App, state)
			isUpdate := false
			prevVersion := ""

			if ok {
				if app.Version == state.Addons.Apps[_idx].Version {
					log.Success(kubernetesCtx, "Already Installed app", "name", app.Name, "version", app.Version)
					continue
				} else {
					// Delete the App
					isUpdate = true
					prevVersion = state.Addons.Apps[_idx].Version
					if err := deleteApplication(k, state.Addons.Apps[_idx]); err != nil {
						return log.NewError(kubernetesCtx, "Update of the App failed Step Uninstall",
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
					return log.NewError(kubernetesCtx, "Update of the App failed Step Install",
						"app", app.Name,
						"FromVer", prevVersion,
						"ToVer", app.Version,
						"errorMsg", err)
				}
				return log.NewError(kubernetesCtx, "App install failed", "app", app, "Reason", err)
			}
			if isUpdate {
				state.Addons.Apps[_idx].Version = app.Version
			} else {
				state.Addons.Apps = append(state.Addons.Apps, storageTypes.Application{
					Name:    app.Name,
					Version: app.Version,
				})
			}
			if err := k.storageDriver.Write(state); err != nil {
				return err
			}

			if isUpdate {
				log.Success(kubernetesCtx, "Updated the App",
					"app", app.Name,
					"FromVer", prevVersion,
					"ToVer", app.Version)
			}
			log.Success(kubernetesCtx, "Installed Application", "name", app.Name, "version", app.Version, "Success", idx+1, "Total", len(apps))
		}

	case consts.OperationDelete:
		for idx, app := range apps {

			_idx, ok := PresentOrNot(app, App, state)
			if !ok {
				log.Success(kubernetesCtx, "App is not present", "name", app.Name, "version", app.Version)
				continue
			}

			if err := deleteApplication(k, app); err != nil {
				return log.NewError(kubernetesCtx, "App uninstall failed", "app", app, "errorMsg", err)
			}

			_cpyApp := utilities.DeepCopySlice[storageTypes.Application](state.Addons.Apps)
			for _i, _app := range state.Addons.Apps {
				if _i != _idx {
					_cpyApp = append(_cpyApp, _app)
				}
			}
			state.Addons.Apps = _cpyApp
			if err := k.storageDriver.Write(state); err != nil {
				return err
			}

			log.Success(kubernetesCtx, "Uninstalled Application", app.Name, "name", "version", app.Version, "Success", idx+1, "Total", len(apps))
		}
	}

	return nil
}

func installApplication(client *Kubernetes, app storageTypes.Application) error {

	if err := helpers.IsValidKsctlComponentVersion(kubernetesCtx, log, app.Version); err != nil {
		return err
	}

	appStruct, err := GetApps(app.Name, app.Version)
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

		log.Box(kubernetesCtx, "App Details via kubectl", appStruct.KubectlConfig.metadata+"\n"+appStruct.KubectlConfig.postInstall)
	}

	log.Success(kubernetesCtx, "Installed Resource")
	return nil
}

func deleteApplication(client *Kubernetes, app storageTypes.Application) error {

	if err := helpers.IsValidKsctlComponentVersion(kubernetesCtx, log, app.Version); err != nil {
		return err
	}
	appStruct, err := GetApps(app.Name, app.Version)
	if err != nil {
		return err
	}

	switch appStruct.InstallType {

	case InstallHelm:
		if err := deleteHelm(client, appStruct); err != nil {
			return err
		}

	case InstallKubectl:
		if err := deleteKubectl(client, appStruct); err != nil {
			return err
		}

	}

	log.Success(kubernetesCtx, "Uninstalled Resource")
	return nil
}
