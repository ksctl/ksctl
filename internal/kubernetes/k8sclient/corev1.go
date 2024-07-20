package k8sclient

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/pkg/types"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) ConfigMapApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.ConfigMap) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "configmap apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "configmap apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) ConfigMapDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.ConfigMap) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		ConfigMaps(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "configmap delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) ServiceDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Service) error {

	ns := o.Namespace
	err := k.clientset.
		CoreV1().
		Services(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "service delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) ServiceApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Service) error {

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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "service apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "service apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) NamespaceCreate(
	ctx context.Context,
	log types.LoggerFactory,
	ns *corev1.Namespace) error {

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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "namespace create failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "namespace create failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) NamespaceDelete(
	ctx context.Context,
	log types.LoggerFactory,
	ns *corev1.Namespace, wait bool) error {

	if err := k.clientset.
		CoreV1().
		Namespaces().
		Delete(context.Background(), ns.Name, metav1.DeleteOptions{
			GracePeriodSeconds: func() *int64 {
				v := int64(0)
				return &v
			}(),
		}); err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "namespace delete failed", "Reason", err),
		)
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
		ctx,
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
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "Failed to get namespace", "namespace", ns.Name, "err", err),
			), true
		},
		func() error {
			log.Success(ctx, "Namespace is completely deleted", "namespace", ns)
			return nil
		},
		fmt.Sprintf("Namespace still deleting: %s", ns.Name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *K8sClient) SecretDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Secret) error {

	err := k.clientset.
		CoreV1().
		Secrets(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "secret delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) SecretApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Secret) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "secret apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "secret apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) PodReadyWait(
	ctx context.Context,
	log types.LoggerFactory,
	name, namespace string) error {

	expoBackoff := helpers.NewBackOff(
		5*time.Second,
		2,
		int(consts.CounterMaxRetryCount),
	)
	var (
		status *corev1.Pod
	)
	_err := expoBackoff.Run(
		ctx,
		log,
		func() (err error) {
			status, err = k.clientset.
				CoreV1().
				Pods(namespace).
				Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "failed to get", "Reason", err))
			}
			return nil
		},
		func() bool {
			return status.Status.Phase == corev1.PodRunning
		},
		func(err error) (errW error, escalateErr bool) {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "pod get failed", "Reason", err),
			), true
		},
		func() error {
			log.Success(ctx, "pod is running", "name", name)
			return nil
		},
		fmt.Sprintf("pod is not ready %s", name),
	)
	if _err != nil {
		return _err
	}
	return nil
}

func (k *K8sClient) PodApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Pod) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "pod apply failed", "Reason", err),
				)
			}

		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "pod apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) PodDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.Pod) error {

	err := k.clientset.
		CoreV1().
		Pods(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "pod delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) ServiceAccountDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.ServiceAccount) error {

	err := k.clientset.
		CoreV1().
		ServiceAccounts(o.Namespace).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "serviceaccount apply failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) ServiceAccountApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *corev1.ServiceAccount) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "serviceaccount apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "serviceaccount apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) NodesList(
	ctx context.Context,
	log types.LoggerFactory) (*corev1.NodeList, error) {
	v, err := k.clientset.
		CoreV1().
		Nodes().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil,
			ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "list nodes failed", "Reason", err),
			)
	}

	return v, nil
}

func (k *K8sClient) NodeDelete(
	ctx context.Context,
	log types.LoggerFactory,
	nodeName string) error {
	err := k.clientset.
		CoreV1().
		Nodes().
		Delete(context.Background(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "node delete failed", "Reason", err),
		)
	}
	return nil
}
