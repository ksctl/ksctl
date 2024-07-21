package kubernetes

import (
	"fmt"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/internal/kubernetes/stacks"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
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

	//var handlers func(app types.KsctlApp, state *storageTypes.StorageDocument) error
	//
	//if op == consts.OperationCreate {
	//	handlers = k.installApplication
	//} else if op == consts.OperationDelete {
	//	handlers = k.deleteApplication
	//}
	//
	//err := handlers(cni, state)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (k *K8sClusterClient) Applications(
	apps []types.KsctlApp,
	state *storageTypes.StorageDocument,
	op consts.KsctlOperation) error {

	var handlers func(app types.KsctlApp, state *storageTypes.StorageDocument) error

	for _, app := range apps {
		if op == consts.OperationCreate {
			handlers = k.installApplication
		} else if op == consts.OperationDelete {
			handlers = k.deleteApplication
		}

		err := handlers(app, state)
		if err != nil {
			return err
		}
	}

	return nil
}

func getStackManifest(app types.KsctlApp, overriding map[string]map[string]any) (metadata.ApplicationStack, error) {
	convertedOverriding := make(map[metadata.StackComponentID]metadata.ComponentOverriding)

	if overriding != nil { // there are some user overriding
		for k, v := range overriding {
			convertedOverriding[metadata.StackComponentID(k)] = metadata.ComponentOverriding(v)
		}
	}

	appStk, err := stacks.FetchKsctlStack(kubernetesCtx, log, app.StackName)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	stackManifest := appStk(metadata.ApplicationParams{
		ComponentParams: convertedOverriding,
	})

	return stackManifest, nil
}

func getComponentVersionOverriding(componentId string, overriding map[string]map[string]any) string {
	if overriding == nil {
		return ""
	}

	if _overridings, found := overriding[componentId]; found {
		if version, safe := _overridings["version"].(string); safe {
			return version
		} else {
			return ""
		}
	}
	return ""
}

func (k *K8sClusterClient) installApplication(
	app types.KsctlApp,
	state *storageTypes.StorageDocument) error {

	stackManifest, err := getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	idxAppInState, foundInState := PresentOrNot(app, App, state)
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

				componentVersion := getComponentVersionOverriding(string(componentId), app.Overrides)

				_err := k.handleInstallComponent(componentId, component)
				if _err != nil {
					return _err
				}

				stateTypeStack.Components[string(componentId)] = storageTypes.Component{
					Version: componentVersion,
				}
			}
			return nil
		}()
		state.Addons.Apps = append(state.Addons.Apps, stateTypeStack)
	} else {
		appInState := state.Addons.Apps[idxAppInState]

		// now I need to find which component changed
		// 1. component is not present in the state
		// 2. version is different

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]
				componentVersion := getComponentVersionOverriding(string(componentId), app.Overrides)

				if componentInState, found := appInState.Components[string(componentId)]; !found {
					_err := k.handleInstallComponent(componentId, component)
					if _err != nil {
						return _err
					}
					appInState.Components[string(componentId)] = storageTypes.Component{
						Version: componentVersion,
					}
				} else {
					if componentVersion != componentInState.Version {
						// WARN: problems might occur like if the next.component had dependency on this
						_err := k.handleUninstallComponent(componentId, component)
						if _err != nil {
							return _err
						}
						componentInState.Version = componentVersion
					} else {
						// already insync with latest changes
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

	return nil
}

func (k *K8sClusterClient) deleteApplication(state *storageTypes.StorageDocument) error {

	// TODO: do we actually need the app, we can use the state and remove using it

	stackManifest, err := getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	idxAppInState, foundInState := PresentOrNot(app, App, state)
	if foundInState {
		// all the things needs to be deleted
	}

	return nil
}

func (k *K8sClusterClient) handleInstallComponent(
	componentId metadata.StackComponentID,
	component metadata.StackComponent) error {

	var errorFailedToInstall = func(err error) error {
		return ksctlErrors.ErrFailedKsctlComponent.Wrap(
			log.NewError(kubernetesCtx, "failed to install", "component", componentId, "Reason", err.Error()),
		)
	}

	if component.HandlerType == metadata.ComponentTypeKubectl {
		if err := installKubectl(k, component.Kubectl); err != nil {
			return errorFailedToInstall(err)
		}
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
			log.NewError(kubernetesCtx, "failed to uninstall", "component", componentId, "Reason", err.Error()),
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

//
//	switch op {
//	case consts.OperationCreate:
//		_, ok := PresentOrNot(cni, Cni, state)
//		if ok {
//			// TODO: need to verify the version based on each component
//			if cni.Version == state.Addons.Cni.Components[0].Version {
//				log.Success(kubernetesCtx, "Already Installed cni", "name", cni.Name, "version", cni.Version)
//				return nil
//			} else {
//				if k.inCluster {
//					return ksctlErrors.ErrInvalidKsctlComponentVersion.Wrap(
//						log.NewError(kubernetesCtx, "We cannot install CNI due to Operation inside the cluster", "name", cni.Name, "version", cni.Version),
//					)
//				} else {
//					log.Box(kubernetesCtx, "Current Impl. doesn't support cni upgrade", `
//Upgrade of CNI is not Possible as of now!
//Reason: if the cni is uninstalled it will lead to all pod in Pending mode
//thus we can't install cni without the help of state.
//So what we can do is Delete it and then
//
//solution is instead of performing k operation inside the cluster which will become hostile
//will perform k only from outside like the ksctl core for the cli or UI
//so what we can do is we can tell ksctl core to fetch latest state and then we can perform operations
//
//another nice thing would be to reconcile every 2 or 5 minutes from the kubernetes cluster Export()
//	(Only k problem will occur for local based system)
//advisiable to use external storage solution
//`)
//				}
//
//				// Step 1: delete the current install
//				// Step 2: install with the new Version
//				return nil // saftey return to avoid version conflicts, k feature is yet to be developed
//			}
//		}
//
//		if err := k.installApplication(cni); err != nil {
//			return err
//		}
//		state.Addons.Cni.Name = cni.StackName
//		state.Addons.Cni.Version = cni.Version
//
//		if err := k.storageDriver.Write(state); err != nil {
//			return err
//		}
//
//		log.Success(kubernetesCtx, "Installed Cni", "name", cni.Name, "version", cni.Version)
//
//	case consts.OperationDelete:
//
//		_, ok := PresentOrNot(cni, App, state)
//		if !ok {
//			log.Success(kubernetesCtx, "Cni is not present", "name", cni.Name, "version", cni.Version)
//			return nil
//		}
//
//		if err := k.deleteApplication(cni); err != nil {
//			return err
//		}
//
//		state.Addons.Cni.Name = ""
//		state.Addons.Cni.Version = ""
//
//		if err := k.storageDriver.Write(state); err != nil {
//			return err
//		}
//
//		log.Success(kubernetesCtx, "Uninstalled Cni", "name", cni.Name, "version", cni.Version)
//	}
//
//	return nil
//}
//
//// Applications Important the sequence of the apps in the list are important
//// it executes from left to right one at a time
//// if it fails at any point of time it stop further installations
//func (k *K8sClusterClient) Applications(apps []types.KsctlApp, state *storageTypes.StorageDocument, op consts.KsctlOperation) error {
//
//	switch op {
//	case consts.OperationCreate:
//		for idx, app := range apps {
//			_idx, foundInKsctlState := PresentOrNot(app, App, state)
//			isUpdate := false
//			prevVersion := ""
//			newVersion := "latest"
//
//			appStk, err := stacks.FetchKsctlStack(kubernetesCtx, log, app.StackName)
//			if err != nil {
//				return err
//			}
//
//			_app := toKsctlApplicationType(app.Overrides)
//
//			stkData := appStk(metadata.ApplicationParams{
//				ComponentParams: _app,
//			})
//
//			// if err := helpers.IsValidKsctlComponentVersion(kubernetesCtx, log, app.Version); err != nil {
//			// 	return err
//			// }
//
//			var changedComponentIDs []metadata.StackComponentID
//			if foundInKsctlState {
//				appInState := state.Addons.Apps[_idx]
//				for _, componentId := range stkData.StkDepsIdx {
//					newRequirements := app.Overrides
//					if v, ok := newRequirements[string(componentId)]; ok {
//						if ver, ok := v["version"]; ok {
//							newVersion = ver.(string)
//						}
//					}
//					if v, isComponentAlreadyPresent := appInState.Components[string(componentId)]; isComponentAlreadyPresent {
//						if v.Version != newVersion {
//							prevVersion = state.Addons.Apps[_idx].Components[string(componentId)].Version
//							changedComponentIDs = append(changedComponentIDs, componentId)
//						}
//					} else {
//						changedComponentIDs = append(changedComponentIDs, componentId)
//					}
//				}
//
//				if len(changedComponentIDs) == 0 {
//					log.Success(kubernetesCtx, "Already Installed app", "name", app.Name, "version", app.Version)
//					continue
//				}
//
//				isUpdate = true
//				if err := k.addOrRemoveApplication(
//					stkData,
//					changedComponentIDs,
//					consts.OperationDelete,
//				); err != nil {
//					if v, ok := err.(ksctlErrors.KsctlWrappedError); ok {
//						return log.NewError(kubernetesCtx, "Update of the App failed Step Uninstall",
//							"app", app.Name,
//							"FromVer", prevVersion,
//							"ToVer", app.Version,
//							"errorMsg", v.Error())
//					}
//					return err
//				}
//
//				// Install the App in the same flow (allowing to flow down)
//			} else {
//				// no data in storageDriver
//				changedComponentIDs = utilities.DeepCopySlice(stkData.StkDepsIdx)
//			}
//
//			if err := k.addOrRemoveApplication(
//				stkData,
//				changedComponentIDs,
//				consts.OperationCreate,
//			); err != nil {
//				if isUpdate {
//					if v, ok := err.(ksctlErrors.KsctlWrappedError); ok {
//						return log.NewError(kubernetesCtx, "Update of the App failed Step Install",
//							"app", app.Name,
//							"FromVer", prevVersion,
//							"ToVer", app.Version,
//							"errorMsg", v.Error())
//					}
//					return err
//				}
//				return err
//			}
//			if isUpdate {
//				state.Addons.Apps[_idx].Version = app.Version
//			} else {
//				state.Addons.Apps = append(state.Addons.Apps, storageTypes.Application{
//					Name:    app.Name,
//					Version: app.Version,
//				})
//			}
//			if err := k.storageDriver.Write(state); err != nil {
//				return err
//			}
//
//			if isUpdate {
//				log.Success(kubernetesCtx, "Updated the App",
//					"app", app.Name,
//					"FromVer", prevVersion,
//					"ToVer", app.Version)
//			}
//			log.Success(kubernetesCtx, "Installed Application", "name", app.Name, "version", app.Version, "Success", idx+1, "Total", len(apps))
//		}
//
//	case consts.OperationDelete:
//		for idx, app := range apps {
//
//			_idx, ok := PresentOrNot(app, App, state)
//			if !ok {
//				log.Success(kubernetesCtx, "App is not present", "name", app.Name, "version", app.Version)
//				continue
//			}
//
//			if err := k.deleteApplication(app); err != nil {
//				return err
//			}
//
//			_cpyApp := utilities.DeepCopySlice[storageTypes.Application](state.Addons.Apps)
//			for _i, _app := range state.Addons.Apps {
//				if _i != _idx {
//					_cpyApp = append(_cpyApp, _app)
//				}
//			}
//			state.Addons.Apps = _cpyApp
//			if err := k.storageDriver.Write(state); err != nil {
//				return err
//			}
//
//			log.Success(kubernetesCtx, "Uninstalled Application", app.Name, "name", "version", app.Version, "Success", idx+1, "Total", len(apps))
//		}
//	}
//
//	return nil
//}
//
//
//// make sure when deleteing you pass the entire components
//func (client *K8sClusterClient) addOrRemoveApplication(
//	stkData metadata.ApplicationStack,
//	changedComponentIDs []metadata.StackComponentID,
//	op consts.KsctlOperation,
//) error {
//	for _, componentId := range stkData.StkDepsIdx {
//		if !utilities.Contains[metadata.StackComponentID](changedComponentIDs, componentId) {
//			log.Print(kubernetesCtx, "Skipping Component", "component", componentId, "Reason no changes detected")
//		} else {
//			component := stkData.Components[componentId]
//
//			switch component.HandlerType {
//			case metadata.ComponentTypeHelm:
//
//				if op == consts.OperationCreate {
//					if err := installHelm(client, component.Helm); err != nil {
//						return ksctlErrors.ErrFailedKsctlComponent.Wrap(
//							log.NewError(kubernetesCtx, "App install failed", "stack", stkData.StackNameID, "component", componentId, "Reason", err.Error()),
//						)
//					}
//				} else if op == consts.OperationDelete {
//					if err := deleteHelm(client, component.Helm); err != nil {
//						return ksctlErrors.ErrFailedKsctlComponent.Wrap(
//							log.NewError(kubernetesCtx, "App delete failed", "stack", stkData.StackNameID, "component", componentId, "Reason", err.Error()),
//						)
//					}
//				}
//
//			case metadata.ComponentTypeKubectl:
//
//				if op == consts.OperationCreate {
//					if err := installKubectl(client, component.Kubectl); err != nil {
//						return ksctlErrors.ErrFailedKsctlComponent.Wrap(
//							log.NewError(kubernetesCtx, "App install failed", "stack", stkData.StackNameID, "component", componentId, "Reason", err.Error()),
//						)
//					}
//					log.Box(kubernetesCtx, "Component Details via kubectl", component.Kubectl.Metadata+"\n"+component.Kubectl.PostInstall)
//				} else if op == consts.OperationDelete {
//					if err := deleteKubectl(client, component.Kubectl); err != nil {
//						return ksctlErrors.ErrFailedKsctlComponent.Wrap(
//							log.NewError(kubernetesCtx, "App delete failed", "stack", stkData.StackNameID, "component", componentId, "Reason", err.Error()),
//						)
//					}
//				}
//
//			}
//		}
//	}
//	if op == consts.OperationCreate {
//		log.Success(kubernetesCtx, "Installed Resource", "stackId", stkData.StackNameID)
//	} else if op == consts.OperationDelete {
//		log.Success(kubernetesCtx, "Uninstalled Resource", "stackId", stkData.StackNameID)
//	}
//
//	return nil
//}

// func (client *K8sClusterClient) installApplication(app types.KsctlApp) error {
//
// 	// TODO: is it even worth it as we are now using map of overridings
// 	// may be we should use https://github.com/go-playground/validator, still need to conform
// 	//if err := helpers.IsValidKsctlComponentVersion(kubernetesCtx, log, app.Version); err != nil {
// 	//	return err
// 	//}
//
// 	appStk, err := stacks.FetchKsctlStack(kubernetesCtx, log, app.StackName)
// 	if err != nil {
// 		return err
// 	}
// 	// WARN: Here we have the components the app is constuated of
//
// 	_app := toKsctlApplicationType(app.Overrides)
//
// 	stkData := appStk(metadata.ApplicationParams{
// 		ComponentParams: _app,
// 	})
//
// 	for _, componentId := range stkData.StkDepsIdx {
// 		component := stkData.Components[componentId]
//
// 		switch component.HandlerType {
// 		case metadata.ComponentTypeHelm:
// 			if err := installHelm(client, component.Helm); err != nil {
// 				return ksctlErrors.ErrFailedKsctlComponent.Wrap(
// 					log.NewError(kubernetesCtx, "App install failed", "app", app, "Reason", err.Error()),
// 				)
// 			}
//
// 		case metadata.ComponentTypeKubectl:
// 			if err := installKubectl(client, component.Kubectl); err != nil {
// 				return ksctlErrors.ErrFailedKsctlComponent.Wrap(
// 					log.NewError(kubernetesCtx, "App install failed", "app", app, "Reason", err.Error()),
// 				)
// 			}
//
// 			log.Box(kubernetesCtx, "App Details via kubectl", component.Kubectl.Metadata+"\n"+component.Kubectl.PostInstall)
// 		}
// 	}
//
// 	log.Success(kubernetesCtx, "Installed Resource", "app", app)
// 	return nil
// }
//
// func (client *K8sClusterClient) deleteApplication(app types.KsctlApp) error {
//
// 	//if err := helpers.IsValidKsctlComponentVersion(kubernetesCtx, log, app.Version); err != nil {
// 	//	return err
// 	//}
// 	appStk, err := stacks.FetchKsctlStack(kubernetesCtx, log, app.StackName)
// 	if err != nil {
// 		return err
// 	}
// 	_app := toKsctlApplicationType(app.Overrides)
//
// 	stkData := appStk(metadata.ApplicationParams{
// 		ComponentParams: _app,
// 	})
//
// 	for _, componentId := range stkData.StkDepsIdx {
// 		component := stkData.Components[componentId]
//
// 		switch component.HandlerType {
// 		case metadata.ComponentTypeHelm:
// 			if err := deleteHelm(client, component.Helm); err != nil {
// 				return ksctlErrors.ErrFailedKsctlComponent.Wrap(
// 					log.NewError(kubernetesCtx, "App delete failed", "app", app, "Reason", err.Error()),
// 				)
// 			}
// 		case metadata.ComponentTypeKubectl:
// 			if err := deleteKubectl(client, component.Kubectl); err != nil {
// 				return ksctlErrors.ErrFailedKsctlComponent.Wrap(
// 					log.NewError(kubernetesCtx, "App delete failed", "app", app, "Reason", err.Error()),
// 				)
// 			}
// 		}
// 	}
//
// 	log.Success(kubernetesCtx, "Uninstalled Resource", "app", app)
// 	return nil
// }
