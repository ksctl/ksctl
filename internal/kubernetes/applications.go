package kubernetes

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/internal/kubernetes/stacks"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

type EnumApplication string

const (
	Cni EnumApplication = "cni"
	App EnumApplication = "app"
)

func PresentOrNot(app types.KsctlApp, typeOfApp EnumApplication, state *storageTypes.StorageDocument) (idx int, isPresent bool) {
	idx = -1

	installedApps := state.Addons

	switch typeOfApp {
	case Cni:
		if app.StackName == installedApps.Cni.Name {
			isPresent = true
			return
		}
	case App:
		for _idx, _app := range installedApps.Apps {
			if _app.Name == app.StackName {
				idx = _idx
				isPresent = true
				return
			}
		}
	}

	return
}

func (k *K8sClusterClient) CNI(
	cni types.KsctlApp,
	state *storageTypes.StorageDocument,
	op consts.KsctlOperation) error {

	var handlers func(app types.KsctlApp, typeOfApp EnumApplication, state *storageTypes.StorageDocument) error

	if op == consts.OperationCreate {
		handlers = k.InstallApplication
	} else if op == consts.OperationDelete {
		handlers = k.deleteApplication
	}

	if op == consts.OperationDelete {
		if err := k.AppPerformPreUninstall(cni, state); err != nil {
			return err
		}
	}

	err := handlers(cni, Cni, state)
	if err != nil {
		return err
	}

	if op == consts.OperationCreate {
		if err := k.AppPerformPostInstall(cni, state); err != nil {
			return err
		}
	}

	return nil
}

func (k *K8sClusterClient) Applications(
	apps []types.KsctlApp,
	state *storageTypes.StorageDocument,
	op consts.KsctlOperation) error {

	var handlers func(app types.KsctlApp, typeOfApp EnumApplication, state *storageTypes.StorageDocument) error

	for _, app := range apps {
		if op == consts.OperationCreate {
			handlers = k.InstallApplication
		} else if op == consts.OperationDelete {
			handlers = k.deleteApplication
		}

		if op == consts.OperationDelete {
			if err := k.AppPerformPreUninstall(app, state); err != nil {
				return err
			}
		}

		err := handlers(app, App, state)
		if err != nil {
			return err
		}

		if op == consts.OperationCreate {
			if err := k.AppPerformPostInstall(app, state); err != nil {
				return err
			}
		}
	}

	return nil
}

func getStackManifest(app types.KsctlApp, overriding map[string]map[string]any) (metadata.ApplicationStack, error) {
	convertedOverriding := make(map[metadata.StackComponentID]metadata.ComponentOverrides)

	if overriding != nil { // there are some user overriding
		for k, v := range overriding {
			convertedOverriding[metadata.StackComponentID(k)] = metadata.ComponentOverrides(v)
		}
	}

	appStk, err := stacks.FetchKsctlStack(kubernetesCtx, log, app.StackName)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	return appStk(metadata.ApplicationParams{
		ComponentParams: convertedOverriding,
	})
}

func getComponentVersionOverriding(component metadata.StackComponent) string {
	if component.HandlerType == metadata.ComponentTypeKubectl {
		return component.Kubectl.Version
	}
	return component.Helm.Charts[0].Version
}

func (k *K8sClusterClient) InstallApplication(
	app types.KsctlApp,
	typeOfApp EnumApplication,
	state *storageTypes.StorageDocument) error {

	stackManifest, err := getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	idxAppInState, foundInState := PresentOrNot(app, typeOfApp, state)
	var (
		errorInStack error
	)
	if !foundInState {
		stateTypeStack := storageTypes.Application{
			Name:       string(stackManifest.StackNameID),
			Components: map[string]storageTypes.Component{},
		}

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := getComponentVersionOverriding(component)

				_err := k.handleInstallComponent(componentId, component)
				if _err != nil {
					return _err
				}

				stateTypeStack.Components[string(componentId)] = storageTypes.Component{
					Version: componentVersion,
				}
				log.Success(
					kubernetesCtx,
					"Installed component",
					"type", typeOfApp,
					"stackId", stackManifest.StackNameID,
					"component", string(componentId),
					"version", componentVersion,
				)
			}
			return nil
		}()
		if typeOfApp == App {
			state.Addons.Apps = append(state.Addons.Apps, stateTypeStack)
		} else {
			state.Addons.Cni = stateTypeStack
		}
	} else {
		var appInState storageTypes.Application
		if typeOfApp == App {
			appInState = state.Addons.Apps[idxAppInState]
		} else {
			appInState = state.Addons.Cni
		}

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := getComponentVersionOverriding(component)

				if componentInState, found := appInState.Components[string(componentId)]; !found {
					_err := k.handleInstallComponent(componentId, component)
					if _err != nil {
						return _err
					}
					appInState.Components[string(componentId)] = storageTypes.Component{
						Version: componentVersion,
					}

					log.Success(
						kubernetesCtx,
						"Installed component",
						"type", typeOfApp,
						"stackId", stackManifest.StackNameID,
						"component", string(componentId),
						"version", componentVersion,
					)
				} else {
					if componentVersion != componentInState.Version {
						// WARN: problems might occur like if the next.component had dependency on this
						if typeOfApp == Cni {
							if k.inCluster {
								return ksctlErrors.ErrInvalidKsctlComponentVersion.Wrap(
									log.NewError(kubernetesCtx, "We cannot install CNI due to Operation inside the cluster", "name", componentId),
								)
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
								return nil
							}
						}

						if _err := k.handleUninstallComponent(componentId, component); _err != nil {
							return log.NewError(
								kubernetesCtx,
								"Update of the App failed Step Uninstall",
								"type", typeOfApp,
								"stackId", stackManifest.StackNameID,
								"component", string(componentId),
								"FromVer", componentInState.Version,
								"ToVer", componentVersion,
								"errorMsg", _err,
							)
						}

						if _err := k.handleInstallComponent(componentId, component); _err != nil {
							return log.NewError(
								kubernetesCtx,
								"Update of the App failed Step Install",
								"type", typeOfApp,
								"stackId", stackManifest.StackNameID,
								"component", string(componentId),
								"FromVer", componentInState.Version,
								"ToVer", componentVersion,
								"errorMsg", _err,
							)
						}
						componentInState.Version = componentVersion
						log.Success(
							kubernetesCtx,
							"Updated the App",
							"type", typeOfApp,
							"stackId", stackManifest.StackNameID,
							"component", string(componentId),
							"FromVer", componentInState.Version,
							"ToVer", componentVersion,
						)

					} else {
						log.Success(kubernetesCtx,
							"Already Installed",
							"type", string(typeOfApp),
							"stackId", stackManifest.StackNameID,
							"component", string(componentId),
							"version", componentVersion,
						)
					}
				}
			}
			return nil
		}()
	}

	if _err := k.storageDriver.Write(state); _err != nil {
		if errorInStack != nil {
			return fmt.Errorf(errorInStack.Error() + " " + _err.Error())
		}
		return _err
	}

	if errorInStack != nil {
		return errorInStack
	}

	log.Success(kubernetesCtx,
		"Installed the Application Stack",
		"stackId", stackManifest.StackNameID,
	)

	return nil
}

func (k *K8sClusterClient) deleteApplication(
	app types.KsctlApp,
	typeOfApp EnumApplication,
	state *storageTypes.StorageDocument) error {

	stackManifest, err := getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	idxAppInState, foundInState := PresentOrNot(app, typeOfApp, state)
	if foundInState {
		errorInStack := func() error {
			for idx := len(stackManifest.StkDepsIdx) - 1; idx >= 0; idx-- {
				componentId := stackManifest.StkDepsIdx[idx]

				component := stackManifest.Components[componentId]
				_err := k.handleUninstallComponent(componentId, component)
				if _err != nil {
					return _err
				}

				log.Success(
					kubernetesCtx,
					"Uninstalled component",
					"type", typeOfApp,
					"stackId", stackManifest.StackNameID,
					"component", string(componentId),
					"version", state.Addons.Cni.Components[string(componentId)].Version,
				)

				if typeOfApp == App {
					delete(state.Addons.Apps[idxAppInState].Components, string(componentId))
				} else {
					delete(state.Addons.Cni.Components, string(componentId))
				}
			}
			return nil
		}()
		if _err := k.storageDriver.Write(state); _err != nil {
			if errorInStack != nil {
				return fmt.Errorf(errorInStack.Error() + " " + _err.Error())
			}
			return _err
		}
		if errorInStack != nil {
			return errorInStack
		}

		if typeOfApp == App {
			leftSide := utilities.DeepCopySlice[storageTypes.Application](state.Addons.Apps[:idxAppInState])
			rightSide := utilities.DeepCopySlice[storageTypes.Application](state.Addons.Apps[idxAppInState+1:])
			state.Addons.Apps = utilities.DeepCopySlice(append(leftSide, rightSide...))
		} else {
			state.Addons.Cni = storageTypes.Application{}
		}
		if _err := k.storageDriver.Write(state); _err != nil {
			return _err
		}
		log.Success(kubernetesCtx,
			"Uninstalled",
			"type", typeOfApp,
			"stackId", stackManifest.StackNameID)
	} else {
		log.Success(
			kubernetesCtx,
			"Already Uninstalled",
			"type", typeOfApp,
			"stackId", stackManifest.StackNameID,
		)
	}
	return nil
}

func (k *K8sClusterClient) handleInstallComponent(
	componentId metadata.StackComponentID,
	component metadata.StackComponent) error {

	var errorFailedToInstall = func(err error) error {
		return ksctlErrors.ErrFailedKsctlComponent.Wrap(
			log.NewError(kubernetesCtx, "failed to install", "component", componentId, "Reason", err),
		)
	}

	if component.HandlerType == metadata.ComponentTypeKubectl {
		if err := installKubectl(k, component.Kubectl); err != nil {
			return errorFailedToInstall(err)
		}
		log.Box(
			kubernetesCtx,
			"Component Details via kubectl",
			component.Kubectl.Metadata+"\n"+component.Kubectl.PostInstall,
		)
	} else {
		if err := installHelm(k, component.Helm); err != nil {
			return errorFailedToInstall(err)
		}
	}
	return nil
}

func (k *K8sClusterClient) handleUninstallComponent(
	componentId metadata.StackComponentID,
	component metadata.StackComponent) error {

	var errorFailedToUninstall = func(err error) error {
		return ksctlErrors.ErrFailedKsctlComponent.Wrap(
			log.NewError(kubernetesCtx, "failed to uninstall", "component", componentId, "Reason", err),
		)
	}

	if component.HandlerType == metadata.ComponentTypeKubectl {
		if err := deleteKubectl(k, component.Kubectl); err != nil {
			return errorFailedToUninstall(err)
		}
	} else {
		if err := deleteHelm(k, component.Helm); err != nil {
			return errorFailedToUninstall(err)
		}
	}
	return nil
}
