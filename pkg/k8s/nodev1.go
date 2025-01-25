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

	nodev1 "k8s.io/api/node/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (k *Client) RuntimeApply(o *nodev1.RuntimeClass,
) error {
	_, err := k.clientset.NodeV1().RuntimeClasses().Create(
		context.Background(),
		o,
		metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.NodeV1().RuntimeClasses().Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "runtime class apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "runtime class apply failed", "Reason", err),
			)
		}
	}

	return nil
}

func (k *Client) RuntimeDelete(resName string) error {
	err := k.clientset.NodeV1().RuntimeClasses().Delete(
		context.Background(),
		resName,
		metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "runtime class delete failed", "Reason", err),
		)
	}

	return nil
}
