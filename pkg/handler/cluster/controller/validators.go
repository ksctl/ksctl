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

package controller

import (
	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/validation"
)

func (cc *Controller) ValidateName(name string) error {
	if err := validation.IsValidName(cc.ctx, cc.l, name); err != nil {
		return err
	}
	return nil
}

func (cc *Controller) ValidateClusterType(clusterType consts.KsctlClusterType) error {
	if !validation.ValidateClusterType(clusterType) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidClusterType,
			cc.l.NewError(
				cc.ctx, "Problem in validation", "clusterType", clusterType,
			),
		)
	}
	return nil
}

func (cc *Controller) ValidateMetadata(c *Client) error {
	meta := c.Metadata
	if _, ok := config.IsContextPresent(cc.KsctlWorkloadConf.WorkerCtx, consts.KsctlContextUser); !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			cc.l.NewError(cc.ctx, "invalid format for context value `USERID`", "Reason", "Make sure the value", "type", "string", "format", `^[\w-]+$`),
		)
	}
	if !validation.ValidateStorage(c.Metadata.StateLocation) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidStorageProvider,
			cc.l.NewError(
				cc.ctx, "Problem in validation", "storage", c.Metadata.StateLocation,
			),
		)
	}

	if !validation.ValidateCloud(meta.Provider) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			cc.l.NewError(
				cc.ctx, "Problem in validation", "cloud", meta.Provider,
			),
		)
	}

	return nil
}
