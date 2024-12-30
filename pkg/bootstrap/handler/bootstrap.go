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
	"github.com/ksctl/ksctl/pkg/provider"
	"sync"

	"github.com/ksctl/ksctl/pkg/bootstrap"
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"

	k3sPkg "github.com/ksctl/ksctl/pkg/bootstrap/distributions/k3s"
	kubeadmPkg "github.com/ksctl/ksctl/pkg/bootstrap/distributions/kubeadm"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
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

	if controllerPayload.Metadata.IsHA {
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

	if kc.s.BootstrapProvider == consts.K8sK3s || kc.s.BootstrapProvider == consts.K8sKubeadm {
		switch kc.s.BootstrapProvider {
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
		}
		return nil
	}

	switch kc.p.Metadata.K8sDistro {
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

			err := kc.p.PreBootstrap.ConfigureDataStore(i)
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

	externalCNI := kc.p.Bootstrap.CNI(kc.p.Metadata.CNIPlugin.StackName)

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
		if len(kc.p.Metadata.CNIPlugin.StackName) == 0 {
			kc.p.Metadata.CNIPlugin.StackName = "flannel"
			kc.p.Metadata.CNIPlugin.Overrides = nil
		}

		if err := k.CNI(kc.p.Metadata.CNIPlugin, kc.s, consts.OperationCreate); err != nil {
			return err
		}

		kc.l.Success(kc.ctx, "Done with installing k8s cni")
	}

	kc.l.Success(kc.ctx, "Done with installing additional k8s tools")
	return nil
}

// TODO: we need to delete any infrastructure resources before we remove this
// Goal: is to trigger a event in the cluster and wait for it
// the event handler will deprovision all the resources before it is safe for us to do ahead
// Warn: this is a deletion procedure
func (kc *Controller) InvokeDestroyProcedure() error {
	kc.l.Success(kc.ctx, "We need to implement the destroy procedure")

	return nil
}

// func handleCreds(ctx context.Context, log logger.Logger, store consts.KsctlStore) (map[string][]byte, error) {
// 	switch store {
// 	case consts.StoreLocal, consts.StoreK8s:
// 		return nil, ksctlErrors.WrapError(
// 			ksctlErrors.ErrInvalidStorageProvider,
// 			log.NewError(ctx, "these are not external storageProvider"),
// 		)
// 	case consts.StoreExtMongo:
// 		return mongodb.ExportEndpoint()
// 	default:
// 		return nil, ksctlErrors.WrapError(
// 			ksctlErrors.ErrInvalidStorageProvider,
// 			log.NewError(ctx, "invalid storage", "storage", store),
// 		)
// 	}
// }
