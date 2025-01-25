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
	"fmt"
	"time"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/waiter"

	"github.com/ksctl/ksctl/pkg/consts"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
)

func (k *Client) DaemonsetApply(o *appsv1.DaemonSet) error {
	ns := o.Namespace

	_, err := k.clientset.
		AppsV1().
		DaemonSets(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				AppsV1().
				DaemonSets(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "daemonset apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "daemonset apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) DeploymentApply(o *appsv1.Deployment) error {
	ns := o.Namespace

	_, err := k.clientset.
		AppsV1().
		Deployments(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				AppsV1().
				Deployments(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "deployment apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "deployment apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) DeploymentReadyWait(name, namespace string) error {

	expoBackoff := waiter.NewWaiter(
		5*time.Second,
		2,
		int(consts.CounterMaxRetryCount),
	)
	var (
		status *appsv1.Deployment
	)
	_err := expoBackoff.Run(
		k.ctx,
		k.l,
		func() (err error) {
			status, err = k.clientset.
				AppsV1().
				Deployments(namespace).
				Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "failed to get", "Reason", err))
			}
			return nil
		},
		func() bool {
			return status.Status.ReadyReplicas > 0
		},
		func(err error) (errW error, escalateErr bool) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "deployment get failed", "Reason", err),
			), true
		},
		func() error {
			k.l.Success(k.ctx, "Few of the replica are ready", "readyReplicas", status.Status.ReadyReplicas)
			return nil
		},
		fmt.Sprintf("retrying no of ready replicas == 0 %s", name),
	)
	if _err != nil {
		return _err
	}

	return nil
}

func (k *Client) DaemonsetDelete(o *appsv1.DaemonSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		DaemonSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "daemonset delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) DeploymentDelete(o *appsv1.Deployment) error {

	ns := o.Namespace
	err := k.clientset.
		AppsV1().
		Deployments(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "deployment delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) StatefulSetApply(o *appsv1.StatefulSet) error {
	ns := o.Namespace

	_, err := k.clientset.
		AppsV1().
		StatefulSets(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				AppsV1().
				StatefulSets(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "statefulset apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "statefulset apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) StatefulSetDelete(o *appsv1.StatefulSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		StatefulSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "statefulset delete failed", "Reason", err),
		)
	}
	return nil
}
