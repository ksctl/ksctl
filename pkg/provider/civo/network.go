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

package civo

import (
	"fmt"
	"github.com/ksctl/ksctl/pkg/waiter"
	"time"

	"github.com/civo/civogo"

	"github.com/ksctl/ksctl/pkg/consts"
)

func (p *Provider) NewNetwork() error {
	name := <-p.chResName

	p.l.Debug(p.ctx, "Printing", "Name", name)

	if len(p.state.CloudInfra.Civo.NetworkID) != 0 {
		p.l.Print(p.ctx, "skipped network creation found", "networkID", p.state.CloudInfra.Civo.NetworkID)
		return nil
	}

	res, err := p.client.CreateNetwork(name)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Civo.NetworkID = res.ID

	net, err := p.client.GetNetwork(res.ID)
	if err != nil {
		return err
	}
	p.state.CloudInfra.Civo.NetworkCIDR = net.CIDR

	p.l.Debug(p.ctx, "Printing", "networkID", res.ID)
	p.l.Debug(p.ctx, "Printing", "networkCIDR", net.CIDR)
	p.l.Success(p.ctx, "Created network", "name", name)

	if err := p.store.Write(p.state); err != nil {
		return err
	}

	expoBackoff := waiter.NewWaiter(
		10*time.Second,
		2,
		int(consts.CounterMaxWatchRetryCount),
	)
	var netInst *civogo.Network
	_err := expoBackoff.Run(
		p.ctx,
		p.l,
		func() (err error) {
			netInst, err = p.client.GetNetwork(res.ID)
			return err
		},
		func() bool {
			return netInst.Status == "Active"
		},
		nil,
		func() error {
			p.l.Print(p.ctx, "network ready", "name", name)
			return nil
		},
		fmt.Sprintf("Waiting for the network %s to be ready", name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (p *Provider) DelNetwork() error {

	if len(p.state.CloudInfra.Civo.NetworkID) == 0 {
		p.l.Print(p.ctx, "skipped network already deleted")
		return nil
	}
	netId := p.state.CloudInfra.Civo.NetworkID

	expoBackoff := waiter.NewWaiter(
		5*time.Second,
		2,
		int(consts.CounterMaxWatchRetryCount),
	)
	_err := expoBackoff.Run(
		p.ctx,
		p.l,
		func() (err error) {
			_, err = p.client.DeleteNetwork(p.state.CloudInfra.Civo.NetworkID)
			return err
		},
		func() bool {
			return true
		},
		nil,
		func() error {
			p.state.CloudInfra.Civo.NetworkID = ""
			return p.store.Write(p.state)
		},
		fmt.Sprintf("Waiting for the network %s to be deleted", p.state.CloudInfra.Civo.NetworkID),
	)
	if _err != nil {
		return _err
	}

	p.l.Success(p.ctx, "Deleted network", "networkID", netId)

	return p.store.DeleteCluster()
}
