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
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/provider"
)

func (kc *Controller) DeleteHACluster() error {

	var err error

	noCP, err := kc.p.Cloud.NoOfControlPlane(kc.p.Metadata.NoCP, false)
	if err != nil {
		return err
	}

	noWP, err := kc.p.Cloud.NoOfWorkerPlane(kc.p.Metadata.NoWP, false)
	if err != nil {
		return err
	}

	noDS, err := kc.p.Cloud.NoOfDataStore(kc.p.Metadata.NoDS, false)
	if err != nil {
		return err
	}

	//////
	wg := &sync.WaitGroup{}
	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, noDS)
	errChanCP := make(chan error, noCP)
	errChanWP := make(chan error, noWP)

	wg.Add(1 + noDS + noCP + noWP)
	//////
	for no := 0; no < noWP; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Role(consts.RoleWp).DelVM(no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	for no := 0; no < noCP; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Role(consts.RoleCp).DelVM(no)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}
	for no := 0; no < noDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Role(consts.RoleDs).DelVM(no)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}

	go func() {
		defer wg.Done()

		err := kc.p.Cloud.Role(consts.RoleLb).DelVM(0)
		if err != nil {
			errChanLB <- err
		}
	}()

	////////
	wg.Wait()
	close(errChanDS)
	close(errChanLB)
	close(errChanCP)
	close(errChanWP)

	for err := range errChanLB {
		if err != nil {
			return err
		}
	}
	for err := range errChanDS {
		if err != nil {
			return err
		}
	}
	for err := range errChanCP {
		if err != nil {
			return err
		}
	}
	for err := range errChanWP {
		if err != nil {
			return err
		}
	}

	pauseOperation(20) // NOTE: experimental time to wait for generic cloud to update its state

	err = kc.p.Cloud.Role(consts.RoleDs).DelFirewall()
	if err != nil {
		return err
	}

	err = kc.p.Cloud.Role(consts.RoleCp).DelFirewall()
	if err != nil {
		return err
	}

	err = kc.p.Cloud.Role(consts.RoleWp).DelFirewall()
	if err != nil {
		return err
	}

	err = kc.p.Cloud.Role(consts.RoleLb).DelFirewall()
	if err != nil {
		return err
	}

	err = kc.p.Cloud.DelSSHKeyPair()
	if err != nil {
		return err
	}

	// NOTE: last one to delete is network
	err = kc.p.Cloud.DelNetwork()
	if err != nil {
		return err
	}

	return nil
}

// AddWorkerNodes the user provides the desired no of workerplane not the no of workerplanes to be added
func (kc *Controller) AddWorkerNodes() (*provider.CloudResourceState, int, error) {

	var err error
	currWP, err := kc.p.Cloud.NoOfWorkerPlane(kc.p.Metadata.NoWP, false)
	if err != nil {
		return nil, -1, err
	}

	_, err = kc.p.Cloud.NoOfWorkerPlane(kc.p.Metadata.NoWP, true)
	if err != nil {
		return nil, -1, err
	}

	wg := &sync.WaitGroup{}

	errChanWP := make(chan error, kc.p.Metadata.NoWP-currWP)

	for no := currWP; no < kc.p.Metadata.NoWP; no++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", kc.p.Metadata.ClusterName, no)).
				Role(consts.RoleWp).
				VMType(kc.p.Metadata.WorkerPlaneNodeType).
				Visibility(true).
				NewVM(no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChanWP)

	for err := range errChanWP {
		if err != nil {
			return nil, -1, err
		}
	}

	transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
	if errState != nil {
		kc.l.Error("handled error", "catch", errState)
		return nil, -1, err
	}

	return &transferableInfraState, currWP, nil
}

// DelWorkerNodes uses the noWP as the desired count of workerplane which is desired
func (kc *Controller) DelWorkerNodes() (*provider.CloudResourceState, []string, error) {

	hostnames := kc.p.Cloud.GetHostNameAllWorkerNode()

	if hostnames == nil {
		return nil, nil, kc.l.NewError(kc.ctx, "hostname is empty")
	}

	currLen := len(hostnames)
	desiredLen := kc.p.Metadata.NoWP
	hostnames = hostnames[desiredLen:]

	if desiredLen < 0 || desiredLen > currLen {
		return nil, nil, kc.l.NewError(kc.ctx, "not a valid count of wp for down scaling")
	}

	wg := &sync.WaitGroup{}
	errChanWP := make(chan error, currLen-desiredLen)

	for no := desiredLen; no < currLen; no++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Role(consts.RoleWp).DelVM(no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}
	wg.Wait()
	close(errChanWP)

	for err := range errChanWP {
		if err != nil {
			return nil, nil, err
		}
	}

	_, err := kc.p.Cloud.NoOfWorkerPlane(desiredLen, true)
	if err != nil {
		return nil, nil, err
	}

	transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
	if errState != nil {
		kc.l.Error("handled error", "catch", errState)
		return nil, nil, err
	}

	return &transferableInfraState, hostnames, nil
}

func (kc *Controller) CreateHACluster() (*provider.CloudResourceState, error) {
	if _, err := kc.p.Cloud.NoOfControlPlane(kc.p.Metadata.NoCP, true); err != nil {
		return nil, err
	}

	if _, err := kc.p.Cloud.NoOfWorkerPlane(kc.p.Metadata.NoWP, true); err != nil {
		return nil, err
	}

	if _, err := kc.p.Cloud.NoOfDataStore(kc.p.Metadata.NoDS, true); err != nil {
		return nil, err
	}

	var err error
	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-net").NewNetwork()
	if err != nil {
		return nil, err
	}

	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-ssh").CreateUploadSSHKeyPair()
	if err != nil {
		return nil, err
	}

	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-fw-lb").
		Role(consts.RoleLb).
		NewFirewall()
	if err != nil {
		return nil, err
	}

	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-fw-db").
		Role(consts.RoleDs).
		NewFirewall()
	if err != nil {
		return nil, err
	}

	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-fw-cp").
		Role(consts.RoleCp).
		NewFirewall()
	if err != nil {
		return nil, err
	}

	err = kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-fw-wp").
		Role(consts.RoleWp).
		NewFirewall()
	if err != nil {
		return nil, err
	}

	//////
	wg := &sync.WaitGroup{}
	errChanLB := make(chan error, 1)
	errChanDS := make(chan error, kc.p.Metadata.NoDS)
	errChanCP := make(chan error, kc.p.Metadata.NoCP)
	errChanWP := make(chan error, kc.p.Metadata.NoWP)

	wg.Add(1 + kc.p.Metadata.NoCP + kc.p.Metadata.NoDS + kc.p.Metadata.NoWP)
	//////
	go func() {
		defer wg.Done()

		err := kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-vm-lb").
			Role(consts.RoleLb).
			VMType(kc.p.Metadata.LoadBalancerNodeType).
			Visibility(true).
			NewVM(0)
		if err != nil {
			errChanLB <- err
		}
	}()

	for no := 0; no < kc.p.Metadata.NoDS; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Name(fmt.Sprintf("%s-vm-db-%d", kc.p.Metadata.ClusterName, no)).
				Role(consts.RoleDs).
				VMType(kc.p.Metadata.DataStoreNodeType).
				Visibility(true).
				NewVM(no)
			if err != nil {
				errChanDS <- err
			}
		}(no)
	}
	for no := 0; no < kc.p.Metadata.NoCP; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Name(fmt.Sprintf("%s-vm-cp-%d", kc.p.Metadata.ClusterName, no)).
				Role(consts.RoleCp).
				VMType(kc.p.Metadata.ControlPlaneNodeType).
				Visibility(true).
				NewVM(no)
			if err != nil {
				errChanCP <- err
			}
		}(no)
	}

	for no := 0; no < kc.p.Metadata.NoWP; no++ {
		go func(no int) {
			defer wg.Done()

			err := kc.p.Cloud.Name(fmt.Sprintf("%s-vm-wp-%d", kc.p.Metadata.ClusterName, no)).
				Role(consts.RoleWp).
				VMType(kc.p.Metadata.WorkerPlaneNodeType).
				Visibility(true).
				NewVM(no)
			if err != nil {
				errChanWP <- err
			}
		}(no)
	}

	////////
	wg.Wait()
	close(errChanDS)
	close(errChanLB)
	close(errChanCP)
	close(errChanWP)

	for err := range errChanLB {
		if err != nil {
			return nil, err
		}
	}
	for err := range errChanDS {
		if err != nil {
			return nil, err
		}
	}
	for err := range errChanCP {
		if err != nil {
			return nil, err
		}
	}
	for err := range errChanWP {
		if err != nil {
			return nil, err
		}
	}

	transferableInfraState, errState := kc.p.Cloud.GetStateForHACluster()
	if errState != nil {
		kc.l.Error("handled error", "catch", errState)
		return nil, err
	}

	return &transferableInfraState, nil
}
