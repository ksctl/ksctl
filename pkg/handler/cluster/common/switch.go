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

package common

import (
	"errors"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/provider/aws"
	"github.com/ksctl/ksctl/v2/pkg/provider/azure"
	"github.com/ksctl/ksctl/v2/pkg/provider/local"
)

func (kc *Controller) Switch() (_ *string, errC error) {
	defer func() {
		if errC != nil {
			v := kc.b.PanicHandler(kc.l)
			if v != nil {
				errC = errors.Join(errC, v)
			}
		}
	}()

	if err := kc.b.ValidateClusterType(kc.p.Metadata.ClusterType); err != nil {
		return nil, err
	}

	if kc.b.IsLocalProvider(kc.p) {
		kc.p.Metadata.Region = "LOCAL"
	}

	if err := kc.p.Storage.Setup(
		kc.p.Metadata.Provider,
		kc.p.Metadata.Region,
		kc.p.Metadata.ClusterName,
		kc.p.Metadata.ClusterType); err != nil {
		return nil, err
	}

	defer func() {
		if err := kc.p.Storage.Kill(); err != nil {
			if errC != nil {
				errC = errors.Join(errC, err)
			} else {
				errC = err
			}
		}
	}()

	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, aws.ProvideClient)
		if err != nil {
			break
		}

	case consts.CloudLocal:
		kc.p.Cloud, err = local.NewClient(kc.ctx, kc.l, kc.p.Metadata, kc.s, kc.p.Storage, local.ProvideClient)

	}

	if err != nil {
		return nil, err
	}

	if errInit := kc.p.Cloud.InitState(consts.OperationGet); errInit != nil {
		return nil, errInit
	}

	if err := kc.p.Cloud.IsPresent(); err != nil {
		return nil, err
	}

	kubeconfig, err := kc.p.Cloud.GetKubeconfig()
	if err != nil {
		return nil, err
	}

	if kubeconfig == nil {
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrKubeconfigOperations,
			kc.l.NewError(
				kc.ctx, "Problem in kubeconfig get, we got nil kubeconfig"),
		)

		return nil, err
	}

	return kubeconfig, nil
}
