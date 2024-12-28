// Copyright 2024 Ksctl Authors
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

package handler

import (
	"context"

	"fmt"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/k8s"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"

	"github.com/ksctl/ksctl/pkg/consts"
)

type (
	StackComponentType uint
	StackComponentID   string
	StackID            string
)

type StackComponent struct {
	Helm        *helm.App
	Kubectl     *k8s.App
	HandlerType StackComponentType
}

// TODO: need to think of taking some sport of the application ksctl provide from the src to some json file in ver control
//
//	so that we can update that and no need of update of the logicial part
//
// also add a String()

type ApplicationStack struct {
	Components map[StackComponentID]StackComponent

	// StkDepsIdx helps you to get sequence of components, aka it acts as a key value table
	StkDepsIdx []StackComponentID

	Maintainer  string
	StackNameID StackID
}

type ApplicationParams struct {
	// StkOverrides   map[string]any
	ComponentParams map[StackComponentID]ComponentOverrides
}

type ComponentOverrides map[string]any

const (
	CiliumStandardStackID  StackID = "cilium"
	FlannelStandardStackID StackID = "flannel"
)

const (
	CiliumComponentID  StackComponentID = "cilium"
	FlannelComponentID StackComponentID = "flannel"
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)

var appsManifests = map[StackID]func(ApplicationParams) (ApplicationStack, error){
	CiliumStandardStackID:  nil,
	FlannelStandardStackID: nil,
}

func FetchKsctlStack(ctx context.Context, log logger.Logger, stkID string) (func(ApplicationParams) (ApplicationStack, error), error) {
	fn, ok := appsManifests[StackID(stkID)]
	if !ok {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			log.NewError(ctx, "appStack not found", "stkId", stkID),
		)
	}
	return fn, nil
}

type EnumApplication string

const (
	Cni EnumApplication = "cni"
	App EnumApplication = "app"
)

func DoesCNIExistOrNot(app controller.KsctlApp, state *statefile.StorageDocument) (isPresent bool) {
	installedApps := state.Addons
	if app.StackName == installedApps.Cni.Name {
		isPresent = true
		return
	}

	return
}

func (k *K8sClusterClient) CNI(
	cni controller.KsctlApp,
	state *statefile.StorageDocument,
	op consts.KsctlOperation) error {

	var handlers func(app controller.KsctlApp, typeOfApp EnumApplication, state *statefile.StorageDocument) error

	if op == consts.OperationCreate {
		handlers = k.InstallApplication
	} else if op == consts.OperationDelete {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Operation not supported for CNI",
		)
	}

	err := handlers(cni, Cni, state)
	if err != nil {
		return err
	}

	return nil
}

func (k *K8sClusterClient) getStackManifest(app controller.KsctlApp, overriding map[string]map[string]any) (ApplicationStack, error) {
	convertedOverriding := make(map[StackComponentID]ComponentOverrides)

	if overriding != nil { // there are some user overriding
		for k, v := range overriding {
			convertedOverriding[StackComponentID(k)] = v
		}
	}

	appStk, err := FetchKsctlStack(k.ctx, k.l, app.StackName)
	if err != nil {
		return ApplicationStack{}, err
	}

	return appStk(ApplicationParams{
		ComponentParams: convertedOverriding,
	})
}

func getComponentVersionOverriding(component StackComponent) string {
	if component.HandlerType == ComponentTypeKubectl {
		return component.Kubectl.Version
	}
	return component.Helm.Charts[0].Version
}

func (k *K8sClusterClient) InstallApplication(
	app controller.KsctlApp,
	typeOfApp EnumApplication,
	state *statefile.StorageDocument) error {

	stackManifest, err := k.getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	foundInState := DoesCNIExistOrNot(app, state)
	var (
		errorInStack error
	)
	if !foundInState {
		stateTypeStack := statefile.Application{
			Name:       string(stackManifest.StackNameID),
			Components: map[string]statefile.Component{},
		}

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := getComponentVersionOverriding(component)

				_err := k.handleInstallComponent(componentId, component)
				if _err != nil {
					return _err
				}

				stateTypeStack.Components[string(componentId)] = statefile.Component{
					Version: componentVersion,
				}
				k.l.Success(
					k.ctx,
					"Installed component",
					"type", typeOfApp,
					"stackId", stackManifest.StackNameID,
					"component", string(componentId),
					"version", componentVersion,
				)
			}
			return nil
		}()
		state.Addons.Cni = stateTypeStack
	} else {
		var appInState statefile.Application
		appInState = state.Addons.Cni

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := getComponentVersionOverriding(component)

				if componentInState, found := appInState.Components[string(componentId)]; !found {
					_err := k.handleInstallComponent(componentId, component)
					if _err != nil {
						return _err
					}
					appInState.Components[string(componentId)] = statefile.Component{
						Version: componentVersion,
					}

					k.l.Success(
						k.ctx,
						"Installed component",
						"type", typeOfApp,
						"stackId", stackManifest.StackNameID,
						"component", string(componentId),
						"version", componentVersion,
					)
				} else {
					if componentVersion != componentInState.Version {
						if typeOfApp == Cni {
							return k.l.NewError(k.ctx, "Current Impl. doesn't support cni upgrade", `
Upgrade of CNI is not Possible as of now!
Reason: We need to add inplace changes!! for helm its possible but for flannel as it uses kubectl not possible
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
					} else {
						k.l.Success(k.ctx,
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

	k.l.Success(k.ctx,
		"Installed the Application Stack",
		"stackId", stackManifest.StackNameID,
	)

	return nil
}

func (k *K8sClusterClient) handleInstallComponent(componentId StackComponentID, component StackComponent) error {
	var errorFailedToInstall = func(err error) error {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlComponent,
			k.l.NewError(k.ctx, "failed to install", "component", componentId, "Reason", err),
		)
	}

	if component.HandlerType == ComponentTypeKubectl {
		if err := k.k8sClient.KubectlApply(component.Kubectl); err != nil {
			return errorFailedToInstall(err)
		}
		k.l.Box(
			k.ctx,
			"Component Details via kubectl",
			component.Kubectl.Metadata+"\n"+component.Kubectl.PostInstall,
		)
	} else {
		if err := k.helmClient.HelmDeploy(component.Helm); err != nil {
			return errorFailedToInstall(err)
		}
	}
	return nil
}
