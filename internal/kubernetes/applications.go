package kubernetes

import (
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

//func (k *K8sClusterClient) InstallCNI(cni types.KsctlApp, state *storageTypes.StorageDocument, op consts.KsctlOperation) error {
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
//func toKsctlApplicationType(from map[string]map[string]any) (to map[metadata.StackComponentID]metadata.ComponentOverriding) {
//	if from == nil {
//		return nil
//	}
//	to = make(map[metadata.StackComponentID]metadata.ComponentOverriding)
//	for k, v := range from {
//		to[metadata.StackComponentID(k)] = metadata.ComponentOverriding(v)
//	}
//	return
//}
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
