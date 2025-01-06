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
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/provider/aws"
	"github.com/ksctl/ksctl/pkg/provider/azure"
)

func (kc *Controller) Credentials() error {
	var err error
	switch kc.p.Metadata.Provider {
	case consts.CloudAzure:
		kc.p.Cloud, err = azure.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, azure.ProvideClient)

	case consts.CloudAws:
		kc.p.Cloud, err = aws.NewClient(kc.ctx, kc.l, kc.p.Metadata, nil, kc.p.Storage, aws.ProvideClient)

	default:
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			kc.l.NewError(
				kc.ctx, "Problem in validation", "cloud", kc.p.Metadata.Provider,
			),
		)
	}

	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}

	err = kc.p.Cloud.Credential()
	if err != nil {
		kc.l.Error("handled error", "catch", err)
		return err
	}
	kc.l.Success(kc.ctx, "Successfully Credential Added")

	return nil
}
