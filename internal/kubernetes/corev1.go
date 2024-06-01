package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) configMapApply(o *corev1.ConfigMap) error {
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
				return log.NewError(kubernetesCtx, "configmap apply failed", "Reason", err)
			}

		} else {
			return log.NewError(kubernetesCtx, "configmap apply failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) configMapDelete(o *corev1.ConfigMap) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		ConfigMaps(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(kubernetesCtx, "configmap delete failed", "Reason", err)
	}
	return nil
}

func (k *Kubernetes) serviceDelete(o *corev1.Service) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		Services(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(kubernetesCtx, "service delete failed", "Reason", err)
	}
	return nil
}

func (k *Kubernetes) serviceApply(o *corev1.Service) error {

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
				return log.NewError(kubernetesCtx, "service apply failed", "Reason", err)
			}

		} else {
			return log.NewError(kubernetesCtx, "service apply failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) namespaceCreate(ns *corev1.Namespace) error {

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
				return log.NewError(kubernetesCtx, "namespace create failed", "Reason", err)
			}
		} else {
			return log.NewError(kubernetesCtx, "namespace create failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) namespaceDelete(ns *corev1.Namespace, wait bool) error {

	if err := k.clientset.
		CoreV1().
		Namespaces().
		Delete(context.Background(), ns.Name, metav1.DeleteOptions{
			GracePeriodSeconds: func() *int64 {
				v := int64(0)
				return &v
			}(),
		}); err != nil {
		return log.NewError(kubernetesCtx, "namespace delete failed", "Reason", err)
	}

	expoBackoff := helpers.NewBackOff(
		10*time.Second,
		1,
		int(consts.CounterMaxRetryCount),
	)
	var (
		errStat error
	)
	_err := expoBackoff.Run(
		kubernetesCtx,
		log,
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
			return log.NewError(kubernetesCtx, "Failed to get namespace", "namespace", ns.Name, "err", err), true
		},
		func() error {
			log.Success(kubernetesCtx, "Namespace is completely deleted", "namespace", ns)
			return nil
		},
		fmt.Sprintf("Namespace still deleting: %s", ns.Name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *Kubernetes) secretDelete(o *corev1.Secret) error {

	err := k.clientset.
		CoreV1().
		Secrets(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(kubernetesCtx, "secret delete failed", "Reason", err)
	}
	return nil
}

func (k *Kubernetes) secretApply(o *corev1.Secret) error {
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
				return log.NewError(kubernetesCtx, "secret apply failed", "Reason", err)
			}

		} else {
			return log.NewError(kubernetesCtx, "secret apply failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) podReadyWait(name, namespace string) error {

	expoBackoff := helpers.NewBackOff(
		5*time.Second,
		2,
		int(consts.CounterMaxRetryCount),
	)
	var (
		status *corev1.Pod
	)
	_err := expoBackoff.Run(
		kubernetesCtx,
		log,
		func() (err error) {
			status, err = k.clientset.
				CoreV1().
				Pods(namespace).
				Get(context.Background(), name, metav1.GetOptions{})
			return err
		},
		func() bool {
			return status.Status.Phase == corev1.PodRunning
		},
		func(err error) (errW error, escalateErr bool) {
			return log.NewError(kubernetesCtx, "pod get failed", "Reason", err), true
		},
		func() error {
			log.Success(kubernetesCtx, "pod is running", "name", name)
			return nil
		},
		fmt.Sprintf("pod is not ready %s", name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *Kubernetes) PodApply(o *corev1.Pod) error {
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
				return log.NewError(kubernetesCtx, "pod apply failed", "Reason", err)
			}

		} else {
			return log.NewError(kubernetesCtx, "pod apply failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) PodDelete(o *corev1.Pod) error {

	err := k.clientset.
		CoreV1().
		Pods(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(kubernetesCtx, "pod delete failed", "Reason", err)
	}
	return nil
}

func (k *Kubernetes) serviceAccountDelete(o *corev1.ServiceAccount) error {

	err := k.clientset.
		CoreV1().
		ServiceAccounts(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(kubernetesCtx, "serviceaccount apply failed", "Reason", err)
	}
	return nil
}

func (k *Kubernetes) serviceAccountApply(o *corev1.ServiceAccount) error {
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
				return log.NewError(kubernetesCtx, "serviceaccount apply failed", "Reason", err)
			}
		} else {
			return log.NewError(kubernetesCtx, "serviceaccount apply failed", "Reason", err)
		}
	}
	return nil
}

func (k *Kubernetes) nodesList() (*corev1.NodeList, error) {
	v, err := k.clientset.
		CoreV1().
		Nodes().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, log.NewError(kubernetesCtx, "list nodes failed", "Reason", err)
	}

	return v, nil
}

func (k *Kubernetes) nodeDelete(nodeName string) error {
	err := k.clientset.
		CoreV1().
		Nodes().
		Delete(context.Background(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(kubernetesCtx, "node delete failed", "Reason", err)
	}
	return nil
}
