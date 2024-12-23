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
	"github.com/ksctl/ksctl/pkg/statefile"
	"os"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
)

func (p *Provider) fetchAPIKey() (string, error) {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken, nil
	}
	p.l.Debug(p.ctx, "environment vars not set: `CIVO_TOKEN`")

	credentials, err := p.store.ReadCredentials(consts.CloudCivo)
	if err != nil {
		return "", err
	}
	if credentials.Civo == nil {
		return "", ksctlErrors.WrapError(
			ksctlErrors.ErrNilCredentials,
			p.l.NewError(p.ctx, "no credentials was found"),
		)
	}
	return credentials.Civo.Token, nil
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

func getValidK8sVersionClient(p *Provider) ([]string, error) {
	vers, err := p.client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil, err
	}
	p.l.Debug(p.ctx, "Printing", "ListAvailableKubernetesVersions", vers)
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(consts.K8sK3s) {
			val = append(val, ver.Label)
		}
	}
	return val, nil
}

func getValidRegionsClient(p *Provider) ([]string, error) {
	regions, err := p.client.ListRegions()
	if err != nil {
		return nil, err
	}
	p.l.Debug(p.ctx, "Printing", "ListRegions", regions)
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val, nil
}

func getValidVMSizesClient(p *Provider) ([]string, error) {
	nodeSizes, err := p.client.ListInstanceSizes()
	if err != nil {
		return nil, err
	}
	p.l.Debug(p.ctx, "Printing", "ListInstanceSizes", nodeSizes)
	var val []string
	for _, region := range nodeSizes {
		val = append(val, region.Name)
	}
	return val, nil
}

func validationOfArguments(p *Provider) error {

	if err := isValidRegion(p, p.Region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(p *Provider, ver string) error {
	valver, err := getValidK8sVersionClient(p)
	if err != nil {
		return err
	}
	for _, vver := range valver {
		if vver == ver {
			return nil
		}
	}
	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidVersion,
		p.l.NewError(p.ctx, "invalid k8s version", "ValidManagedK8sVersions", valver),
	)
}

func isValidRegion(p *Provider, reg string) error {
	validFromClient, err := getValidRegionsClient(p)
	if err != nil {
		return err
	}
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudRegion,
		p.l.NewError(p.ctx, "invalid region", "ValidRegion", validFromClient),
	)
}

func isValidVMSize(p *Provider, size string) error {
	validFromClient, err := getValidVMSizesClient(p)
	if err != nil {
		return err
	}
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return ksctlErrors.WrapError(
		ksctlErrors.ErrInvalidCloudVMSize,
		p.l.NewError(p.ctx, "invalid Virtual Machine size", "ValidVMSize", validFromClient),
	)
}
