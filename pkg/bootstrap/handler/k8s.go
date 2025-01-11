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

	"github.com/ksctl/ksctl/pkg/apps/stack"
	"github.com/ksctl/ksctl/pkg/bootstrap/handler/cni"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/k8s"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type K8sClusterClient struct {
	ctx           context.Context
	l             logger.Logger
	storageDriver storage.Storage

	r *rest.Config

	helmClient *helm.Client
	k8sClient  *k8s.Client
	inCluster  bool
}

func NewClusterClient(
	parentCtx context.Context,
	parentLog logger.Logger,
	storage storage.Storage,
	kubeconfig string,
) (k *K8sClusterClient, err error) {
	k = &K8sClusterClient{
		storageDriver: storage,
	}

	k.ctx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, "kubernetes-client")
	k.l = parentLog

	rawKubeconfig := []byte(kubeconfig)

	config := &rest.Config{}
	config, err = clientcmd.BuildConfigFromKubeconfigGetter(
		"",
		func() (*api.Config, error) {
			return clientcmd.Load(rawKubeconfig)
		})
	if err != nil {
		return
	}

	k.k8sClient, err = k8s.NewK8sClient(parentCtx, parentLog, config)
	if err != nil {
		return
	}
	k.r = config

	k.helmClient, err = helm.NewKubeconfigHelmClient(
		k.ctx,
		k.l,
		kubeconfig,
	)
	if err != nil {
		return
	}

	return k, nil
}

func (k *K8sClusterClient) CNI(
	cni stack.KsctlApp,
	state *statefile.StorageDocument,
	op consts.KsctlOperation) error {

	if op == consts.OperationDelete {
		return ksctlErrors.WrapErrorf(
			ksctlErrors.ErrInternal,
			"Operation not supported for CNI",
		)
	}

	err := k.installCni(cni, state)
	if err != nil {
		return err
	}

	return nil
}

func (k *K8sClusterClient) getStackManifest(
	app stack.KsctlApp,
	overriding map[string]map[string]any,
) (stack.ApplicationStack, error) {

	convertedOverriding := make(map[stack.ComponentID]stack.ComponentOverrides)

	if overriding != nil { // there are some user overriding
		for k, v := range overriding {
			convertedOverriding[stack.ComponentID(k)] = v
		}
	}

	appStk, err := cni.FetchKsctlStack(k.ctx, k.l, app.StackName)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return appStk(stack.ApplicationParams{
		ComponentParams: convertedOverriding,
	})
}

func (k *K8sClusterClient) installCni(
	app stack.KsctlApp,
	state *statefile.StorageDocument) error {

	stackManifest, err := k.getStackManifest(app, app.Overrides)
	if err != nil {
		return err
	}

	foundInState := cni.DoesCNIExistOrNot(app, state)
	var (
		errorInStack error
	)
	if !foundInState {
		stateTypeStack := statefile.SlimProvisionerAddon{
			Name:                    string(stackManifest.StackNameID),
			For:                     consts.K8sKsctl,
			KsctlSpecificComponents: make(map[string]statefile.KsctlSpecificComponent, len(stackManifest.Components)),
		}

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := stack.GetComponentVersionOverriding(component)

				_err := k.handleInstallComponent(componentId, component)
				if _err != nil {
					return _err
				}

				stateTypeStack.KsctlSpecificComponents[string(componentId)] = statefile.KsctlSpecificComponent{
					Version: componentVersion,
				}

				k.l.Success(
					k.ctx,
					"Installed component",
					"type", "cni",
					"stackId", stackManifest.StackNameID,
					"component", string(componentId),
					"version", componentVersion,
				)
			}
			return nil
		}()
		state.ProvisionerAddons.Cni = stateTypeStack
	} else {
		appInState := state.ProvisionerAddons.Cni

		errorInStack = func() error {
			for _, componentId := range stackManifest.StkDepsIdx {
				component := stackManifest.Components[componentId]

				componentVersion := stack.GetComponentVersionOverriding(component)

				if componentInState, found := appInState.KsctlSpecificComponents[string(componentId)]; !found {
					_err := k.handleInstallComponent(componentId, component)
					if _err != nil {
						return _err
					}
					appInState.KsctlSpecificComponents[string(componentId)] = statefile.KsctlSpecificComponent{
						Version: componentVersion,
					}

					k.l.Success(
						k.ctx,
						"Installed component",
						"type", "cni",
						"stackId", stackManifest.StackNameID,
						"component", string(componentId),
						"version", componentVersion,
					)
				} else {
					if componentVersion != componentInState.Version {
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
					} else {
						k.l.Success(k.ctx,
							"Already Installed",
							"type", "cni",
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

func (k *K8sClusterClient) handleInstallComponent(componentId stack.ComponentID, component stack.Component) error {
	var errorFailedToInstall = func(err error) error {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlComponent,
			k.l.NewError(k.ctx, "failed to install", "component", componentId, "Reason", err),
		)
	}

	if component.HandlerType == stack.ComponentTypeKubectl {
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
