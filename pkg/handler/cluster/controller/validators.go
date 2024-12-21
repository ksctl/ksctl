// Copyright 2024 ksctl
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
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/validation"
)

func (manager *Controller) ValidateMetadata(c *Client) error {
	meta := c.Metadata
	if _, ok := config.IsContextPresent(manager.ctx, consts.KsctlContextUserID); !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			manager.l.NewError(manager.ctx, "invalid format for context value `USERID`", "Reason", "Make sure the value", "type", "string", "format", `^[\w-]+$`),
		)
	}

	if !validation.ValidateCloud(meta.Provider) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			manager.l.NewError(
				manager.ctx, "Problem in validation", "cloud", meta.Provider,
			),
		)
	}
	if !validation.ValidateDistro(meta.K8sDistro) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidBootstrapProvider,
			manager.l.NewError(
				manager.ctx, "Problem in validation", "bootstrap", meta.K8sDistro,
			),
		)
	}

	if err := validation.IsValidName(manager.ctx, manager.l, meta.ClusterName); err != nil {
		return err
	}

	return nil
}
