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

package azure

import (
	"fmt"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/statefile"
)

func generateResourceGroupName(clusterName, clusterType string) string {
	return fmt.Sprintf("ksctl-resgrp-%s-%s", clusterType, clusterName)
}

func (p *Provider) loadStateHelper() error {
	raw, err := p.store.Read()
	if err != nil {
		return err
	}
	*p.state = func(x *statefile.StorageDocument) statefile.StorageDocument {
		return *x
	}(raw)
	return nil
}

func (p *Provider) validationOfArguments() error {

	if err := p.isValidRegion(p.Region); err != nil {
		return err
	}

	return nil
}

func (p *Provider) isValidK8sVersion(ver string) error {
	res, err := p.client.ListKubernetesVersions()
	if err != nil {
		return err
	}

	p.l.Debug(p.ctx, "Printing", "ListKubernetesVersions", res)

	var vers []string
	for _, version := range res.Values {
		vers = append(vers, *version.Version)
	}
	for _, valver := range vers {
		if valver == ver {
			return nil
		}
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidVersion,
		p.l.NewError(p.ctx, "invalid k8s version", "ValidManagedK8sVersions", vers),
	)
}

func (p *Provider) isValidRegion(reg string) error {
	validReg, err := p.client.ListLocations()
	if err != nil {
		return err
	}
	p.l.Debug(p.ctx, "Printing", "ListLocation", validReg)

	for _, valid := range validReg {
		if valid == reg {
			return nil
		}
	}
	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudRegion,
		p.l.NewError(p.ctx, "Invalid region", "Valid options", validReg),
	)
}

func (p *Provider) isValidVMSize(size string) error {

	validSize, err := p.client.ListVMTypes()
	if err != nil {
		return err
	}
	p.l.Debug(p.ctx, "Printing", "ListVMType", validSize)

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudVMSize,
		p.l.NewError(p.ctx, "Invalid vm size", "Valid options", validSize),
	)
}
