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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/k8s"
	"github.com/ksctl/ksctl/v2/pkg/provider"
	"github.com/ksctl/ksctl/v2/pkg/utilities"

	"github.com/ksctl/ksctl/v2/pkg/bootstrap"
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"

	k3sPkg "github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions/k3s"
	kubeadmPkg "github.com/ksctl/ksctl/v2/pkg/bootstrap/distributions/kubeadm"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct {
	ctx context.Context
	l   logger.Logger
	p   *controller.Client
	b   *controller.Controller
	s   *statefile.StorageDocument
}

func NewController(
	ctx context.Context,
	log logger.Logger,
	baseController *controller.Controller,
	state *statefile.StorageDocument,
	operation consts.KsctlOperation,
	transferableInfraState *provider.CloudResourceState,
	controllerPayload *controller.Client,
) (*Controller, error) {

	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "pkg/bootstrap/handler/bootstrap")
	cc.l = log
	cc.b = baseController
	cc.p = controllerPayload
	cc.s = state

	if controllerPayload.Metadata.ClusterType == consts.ClusterTypeSelfMang {
		err := cc.setupInterfaces(operation, transferableInfraState)
		if err != nil {
			return nil, err
		}
	}

	return cc, nil
}

func (kc *Controller) setupInterfaces(
	operation consts.KsctlOperation,
	transferableInfraState *provider.CloudResourceState,
) error {

	kc.p.PreBootstrap = bootstrap.NewPreBootStrap(
		kc.ctx,
		kc.l,
		kc.s,
		kc.p.Storage,
	)

	var bootstrapProvider consts.KsctlKubernetes

	if kc.s.BootstrapProvider == consts.K8sK3s || kc.s.BootstrapProvider == consts.K8sKubeadm {
		bootstrapProvider = kc.s.BootstrapProvider
	} else {
		bootstrapProvider = kc.p.Metadata.K8sDistro
	}

	switch bootstrapProvider {
	case consts.K8sK3s:
		kc.p.Bootstrap = k3sPkg.NewClient(
			kc.ctx,
			kc.l,
			kc.p.Storage,
			kc.s,
		)
	case consts.K8sKubeadm:
		kc.p.Bootstrap = kubeadmPkg.NewClient(
			kc.ctx,
			kc.l,
			kc.p.Storage,
			kc.s,
		)
	default:
		return kc.l.NewError(kc.ctx, "Invalid k8s provider")
	}

	if errTransfer := kc.p.PreBootstrap.Setup(transferableInfraState, operation); errTransfer != nil {
		kc.l.Error("handled error", "catch", errTransfer)
		return errTransfer
	}
	return nil
}

func (kc *Controller) ConfigureCluster() (bool, error) {
	waitForPre := &sync.WaitGroup{}

	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, kc.p.Metadata.NoDS)

	waitForPre.Add(1 + kc.p.Metadata.NoDS)

	go func() {
		defer waitForPre.Done()

		err := kc.p.PreBootstrap.ConfigureLoadbalancer()
		if err != nil {
			errChanLB <- err
		}
	}()

	for no := 0; no < kc.p.Metadata.NoDS; no++ {
		go func(i int) {
			defer waitForPre.Done()

			err := kc.p.PreBootstrap.ConfigureDataStore(i, kc.p.Metadata.EtcdVersion)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}
	waitForPre.Wait()
	close(errChanLB)
	close(errChanDS)

	for err := range errChanLB {
		if err != nil {
			return false, err
		}
	}

	for err := range errChanDS {
		if err != nil {
			return false, err
		}
	}

	if err := kc.p.Bootstrap.Setup(consts.OperationCreate); err != nil {
		return false, err
	}

	externalCNI := kc.p.Bootstrap.CNI(kc.p.Metadata.Addons)

	kc.p.Bootstrap = kc.p.Bootstrap.K8sVersion(kc.p.Metadata.K8sVersion)
	if kc.p.Bootstrap == nil {
		return false, kc.l.NewError(kc.ctx, "invalid version of self-managed k8s cluster")
	}

	// wp[0,N] depends on cp[0]
	err := kc.p.Bootstrap.ConfigureControlPlane(0)
	if err != nil {
		return false, err
	}

	errChanCP := make(chan error, kc.p.Metadata.NoCP-1)
	errChanWP := make(chan error, kc.p.Metadata.NoWP)

	wg := &sync.WaitGroup{}

	wg.Add(kc.p.Metadata.NoCP - 1 + kc.p.Metadata.NoWP)

	for no := 1; no < kc.p.Metadata.NoCP; no++ {
		go func(i int) {
			defer wg.Done()
			err := kc.p.Bootstrap.ConfigureControlPlane(i)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}
	for no := 0; no < kc.p.Metadata.NoWP; no++ {
		go func(i int) {
			defer wg.Done()
			err := kc.p.Bootstrap.JoinWorkerplane(i)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()

	close(errChanCP)
	close(errChanWP)

	for err := range errChanCP {
		if err != nil {
			return false, err
		}
	}

	for err := range errChanWP {
		if err != nil {
			return false, err
		}
	}
	return externalCNI, nil
}

func (kc *Controller) JoinMoreWorkerPlanes(start, end int) error {

	if err := kc.p.Bootstrap.Setup(consts.OperationGet); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	errChan := make(chan error, end-start)

	for no := start; no < end; no++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := kc.p.Bootstrap.JoinWorkerplane(i)
			if err != nil {
				errChan <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (kc *Controller) DelWorkerPlanes(kubeconfig string, hostnames []string) error {

	k, err := NewClusterClient(
		kc.ctx,
		kc.l,
		kc.p.Storage,
		kubeconfig,
	)
	if err != nil {
		return err
	}

	for _, hostname := range hostnames {
		if err := k.DeleteWorkerNodes(hostname); err != nil {
			return err
		}
	}
	return nil
}

func (kc *Controller) InstallAdditionalTools(externalCNI bool) error {

	if _, ok := config.IsContextPresent(kc.ctx, consts.KsctlTestFlagKey); ok {
		return nil
	}

	k, err := NewClusterClient(
		kc.ctx,
		kc.l,
		kc.p.Storage,
		kc.s.ClusterKubeConfig,
	)
	if err != nil {
		return err
	}

	if externalCNI {
		kc.l.Print(kc.ctx, "Installing External CNI Plugin")
		_addons := kc.p.Metadata.Addons.GetAddons("ksctl")
		var _cni *addons.ClusterAddon
		for _, addon := range _addons {
			if addon.IsCNI {
				_cni = &addon
				break
			}
		}
		_c := stack.KsctlApp{}

		if _cni == nil {
			kc.l.Print(kc.ctx, "CNI Plugin not found in addons list")

			_c.StackName = "cilium"
			_c.Overrides = nil
		} else {
			_c.StackName = _cni.Name
			if _cni.Config == nil {
				_c.Overrides = nil
			} else {
				_v := map[string]map[string]any{}
				if err := json.Unmarshal([]byte(*_cni.Config), &_v); err != nil {
					return kc.l.NewError(kc.ctx, "failed to deserialize cni config", "error", err)
				}
				_c.Overrides = _v
			}
		}

		if err := k.CNI(_c, kc.s, consts.OperationCreate); err != nil {
			return err
		}

		kc.l.Success(kc.ctx, "Done with installing k8s cni")
	}

	return nil
}

func (kc *Controller) InvokeDestroyProcedure() error {
	kc.l.Debug(kc.ctx, "We need to implement the destroy procedure")

	return nil
}

func (kc *Controller) InstallKcm(
	kubeconfig *string,
	namespace string,
	clusterConfigMapName string,
	app statefile.SlimProvisionerAddon,
) error {
	k, err := NewClusterClient(
		kc.ctx,
		kc.l,
		kc.p.Storage,
		*kubeconfig,
	)
	if err != nil {
		return err
	}

	if err := k.k8sClient.NamespaceCreate(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
	); err != nil {
		kc.l.Error("failed to create namespace", "namespace", namespace, "error", err)
		return err
	}

	serializeVersion := func(versions map[string]*string) []byte {
		versionsBytes, err := json.Marshal(versions)
		if err != nil {
			kc.l.Error("failed to serialize versions", "error", err)
			return nil
		}
		return versionsBytes
	}

	ksctlConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"cluster_type":     kc.s.ClusterType,
			"cluster_name":     kc.s.ClusterName,
			"cluster_region":   kc.s.Region,
			"cloud":            string(kc.s.InfraProvider),
			"k8s_distribution": string(kc.s.BootstrapProvider),
		},
		BinaryData: map[string][]byte{
			"versions": func() []byte {
				versions := map[string]*string{
					"k3s":     kc.s.Versions.K3s,
					"kubeadm": kc.s.Versions.Kubeadm,
					"aks":     kc.s.Versions.Aks,
					"eks":     kc.s.Versions.Eks,
					"haproxy": kc.s.Versions.HAProxy,
					"etcd":    kc.s.Versions.Etcd,
				}

				return serializeVersion(versions)
			}(),
		},
		Immutable: utilities.Ptr(true),
	}

	if err := k.k8sClient.ConfigMapApply(ksctlConfigMap); err != nil {
		kc.l.Error("failed to create configmap", "configmap", clusterConfigMapName, "error", err)
		return err
	}

	if err := k.k8sClient.KubectlApply(&k8s.App{
		Version: *app.Version,
		Urls:    []string{fmt.Sprintf("https://github.com/ksctl/kcm/releases/download/%s/install.yaml", *app.Version)},
	}); err != nil {
		return err
	}

	kc.s.ProvisionerAddons.Apps = append(kc.s.ProvisionerAddons.Apps, app)
	if err := kc.p.Storage.Write(kc.s); err != nil {
		kc.l.Error("failed to write statefile", "error", err)
		return err
	}

	return nil
}

func (kc *Controller) UninstallKcm(
	kubeconfig *string,
	namespace string,
	clusterConfigMapName string,
	app statefile.SlimProvisionerAddon,
) error {
	k, err := NewClusterClient(
		kc.ctx,
		kc.l,
		kc.p.Storage,
		*kubeconfig,
	)
	if err != nil {
		return err
	}

	if err := k.k8sClient.ConfigMapDelete(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterConfigMapName,
				Namespace: namespace,
			},
		},
	); err != nil {
		kc.l.Error("failed to delete configmap", "configmap", clusterConfigMapName, "error", err)
		// return err // ignore this error
	}

	if err := k.k8sClient.NamespaceDelete(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		true,
	); err != nil {
		kc.l.Error("failed to delete namespace", "namespace", namespace, "error", err)
		return err
	}

	idx := -1

	for i, addon := range kc.s.ProvisionerAddons.Apps {
		if addon.Name == app.Name && addon.For == app.For {
			idx = i
			if err := k.k8sClient.KubectlDelete(&k8s.App{
				Version: *addon.Version,
				Urls:    []string{fmt.Sprintf("https://github.com/ksctl/kcm/releases/download/%s/install.yaml", *addon.Version)},
			}); err != nil {
				return err
			}
		}
	}
	if idx != -1 {
		kc.s.ProvisionerAddons.Apps = append(kc.s.ProvisionerAddons.Apps[:idx], kc.s.ProvisionerAddons.Apps[idx+1:]...)
		if err := kc.p.Storage.Write(kc.s); err != nil {
			kc.l.Error("failed to write statefile", "error", err)
			return err
		}
	}

	return nil
}
