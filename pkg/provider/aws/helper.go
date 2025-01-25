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

package aws

import (
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/validation"
)

func (p *Provider) validationOfArguments() error {

	if err := p.isValidRegion(); err != nil {
		return err
	}

	if err := validation.IsValidName(p.ctx, p.l, p.ClusterName); err != nil {
		return err
	}

	return nil
}

func (p *Provider) isValidRegion() error {

	validReg, err := p.client.ListLocations()
	if err != nil {
		return err
	}

	for _, reg := range validReg {
		if reg == p.Region {
			return nil
		}
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudRegion,
		p.l.NewError(p.ctx, "region not found", "validRegions", validReg),
	)
}

func (p *Provider) isValidVMSize(size string) error {
	validSize, err := p.client.ListVMTypes()
	if err != nil {
		return err
	}

	for _, valid := range validSize.InstanceTypes {
		constAsString := string(valid.InstanceType)
		if constAsString == size {
			return nil
		}
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudVMSize,
		p.l.NewError(p.ctx, "invalid vm size", "Valid options", validSize),
	)
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

func (p *Provider) isValidK8sVersion(version string) error {
	validVersions, err := p.client.ListK8sVersions(p.ctx)
	if err != nil {
		return err
	}

	for _, ver := range validVersions {
		if ver == version {
			return nil
		}
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidVersion,
		p.l.NewError(p.ctx, "invalid k8s version", "validVersions", validVersions),
	)
}
