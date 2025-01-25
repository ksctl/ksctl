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

// TODO(@dipankardas011): Please depricate these interfaces
package k8s

import (
	"context"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Client) ApiExtensionsApply(o *apiextensionsv1.CustomResourceDefinition) error {

	_, err := k.apiextensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.apiextensionsClient.
				ApiextensionsV1().
				CustomResourceDefinitions().
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "apiExtension apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "apiExtension apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) ApiExtensionsDelete(o *apiextensionsv1.CustomResourceDefinition) error {

	err := k.apiextensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "apiExtension delete failed", "Reason", err),
		)
	}
	return nil
}
