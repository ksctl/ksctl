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

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/waiter"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Client) ConfigMapApply(o *corev1.ConfigMap) error {
	ns := o.Namespace

	_, err := k.clientset.
		CoreV1().
		ConfigMaps(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = k.clientset.
				CoreV1().
				ConfigMaps(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "configmap apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "configmap apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) ConfigMapDelete(o *corev1.ConfigMap) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		ConfigMaps(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "configmap delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) ServiceDelete(o *corev1.Service) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		Services(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "service delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) ServiceApply(o *corev1.Service) error {

	ns := o.Namespace
	_, err := k.clientset.
		CoreV1().
		Services(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = k.clientset.
				CoreV1().
				Services(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "service apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "service apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) NamespaceCreate(ns *corev1.Namespace) error {

	if _, err := k.clientset.
		CoreV1().
		Namespaces().
		Create(context.Background(), ns, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				CoreV1().
				Namespaces().
				Update(context.Background(), ns, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "namespace create failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "namespace create failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) NamespaceDelete(ns *corev1.Namespace, wait bool) error {

	if err := k.clientset.
		CoreV1().
		Namespaces().
		Delete(context.Background(), ns.Name, metav1.DeleteOptions{
			GracePeriodSeconds: func() *int64 {
				v := int64(0)
				return &v
			}(),
		}); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "namespace delete failed", "Reason", err),
		)
	}

	expoBackoff := waiter.NewWaiter(
		10*time.Second,
		1,
		int(consts.CounterMaxRetryCount),
	)
	var (
		errStat error
	)
	_err := expoBackoff.Run(
		k.ctx,
		k.l,
		func() (err error) {

			_, errStat = k.clientset.
				CoreV1().
				Namespaces().
				Get(context.Background(), ns.Name, metav1.GetOptions{})
			return err
		},
		func() bool {
			return apierrors.IsNotFound(errStat) || apierrors.IsGone(errStat)
		},
		func(err error) (errW error, escalateErr bool) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "Failed to get namespace", "namespace", ns.Name, "err", err),
			), true
		},
		func() error {
			k.l.Success(k.ctx, "Namespace is completely deleted", "namespace", ns)
			return nil
		},
		fmt.Sprintf("Namespace still deleting: %s", ns.Name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *Client) SecretDelete(o *corev1.Secret) error {

	err := k.clientset.
		CoreV1().
		Secrets(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "secret delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) SecretApply(o *corev1.Secret) error {
	ns := o.Namespace

	_, err := k.clientset.
		CoreV1().
		Secrets(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = k.clientset.
				CoreV1().
				Secrets(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "secret apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "secret apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) PodReadyWait(name, namespace string) error {

	expoBackoff := waiter.NewWaiter(
		5*time.Second,
		2,
		int(consts.CounterMaxRetryCount),
	)
	var (
		status *corev1.Pod
	)
	_err := expoBackoff.Run(
		k.ctx,
		k.l,
		func() (err error) {
			status, err = k.clientset.
				CoreV1().
				Pods(namespace).
				Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "failed to get", "Reason", err))
			}
			return nil
		},
		func() bool {
			return status.Status.Phase == corev1.PodRunning
		},
		func(err error) (errW error, escalateErr bool) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "pod get failed", "Reason", err),
			), true
		},
		func() error {
			k.l.Success(k.ctx, "pod is running", "name", name)
			return nil
		},
		fmt.Sprintf("pod is not ready %s", name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *Client) PodApply(o *corev1.Pod) error {
	ns := o.Namespace

	_, err := k.clientset.
		CoreV1().
		Pods(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = k.clientset.
				CoreV1().
				Pods(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "pod apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "pod apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) PodDelete(o *corev1.Pod) error {

	err := k.clientset.
		CoreV1().
		Pods(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{
			GracePeriodSeconds: o.DeletionGracePeriodSeconds,
		})

	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "pod delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) ServiceAccountDelete(o *corev1.ServiceAccount) error {

	err := k.clientset.
		CoreV1().
		ServiceAccounts(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "serviceaccount apply failed", "Reason", err),
		)
	}
	return nil
}

func (k *Client) ServiceAccountApply(o *corev1.ServiceAccount) error {
	ns := o.Namespace

	_, err := k.clientset.
		CoreV1().
		ServiceAccounts(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				CoreV1().
				ServiceAccounts(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "serviceaccount apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "serviceaccount apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) NodesList() (*corev1.NodeList, error) {
	v, err := k.clientset.
		CoreV1().
		Nodes().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "list nodes failed", "Reason", err),
			)
	}

	return v, nil
}

func (k *Client) NodeUpdate(node *corev1.Node,
) (*corev1.Node, error) {
	v, err := k.clientset.
		CoreV1().
		Nodes().
		Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return nil,
			ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "node delete failed", "Reason", err),
			)
	}
	return v, nil
}

func (k *Client) NodeCordon(nodeName string) error {
	node, err := k.clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "node get failed", "Reason", err),
		)
	}

	node.Spec.Unschedulable = true

	_, err = k.clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "node cordon failed", "Reason", err),
		)
	}

	return nil
}

func (k *Client) NodeDrain(nodeName string) error {
	// Refer: https://kubernetes.io/images/docs/kubectl_drain.svg
	if err := k.NodeCordon(nodeName); err != nil {
		return err
	}

	pods, err := k.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "failed to list pods on node", "Reason", err),
		)
	}

	for _, pod := range pods.Items {
		// Skip DaemonSet pods
		if pod.DeletionGracePeriodSeconds != nil {
			k.l.Print(k.ctx, "Skipping pod because it belongs to DaemonSet", "pod", pod.Name)
			continue
		}

		err := k.PodDelete(&pod)
		if err != nil {
			k.l.Warn(k.ctx, "Failed to evict pod", "pod", pod.Name, "err", err)
			continue
		}
		k.l.Success(k.ctx, "Evicted pod", "pod", pod.Name)
	}

	return nil
}

func (k *Client) NodeDelete(nodeName string) error {

	if err := k.NodeDrain(nodeName); err != nil {
		return err
	}

	err := k.clientset.
		CoreV1().
		Nodes().
		Delete(context.Background(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "node delete failed", "Reason", err),
		)
	}
	return nil
}
