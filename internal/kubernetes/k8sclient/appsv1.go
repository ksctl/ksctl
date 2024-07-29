package k8sclient

import (
	"context"
	"fmt"
	"time"

	"github.com/ksctl/ksctl/pkg/types"

	"github.com/ksctl/ksctl/pkg/helpers"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
)

func (k *K8sClient) DaemonsetApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.DaemonSet) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "daemonset apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "daemonset apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) DeploymentApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.Deployment) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "deployment apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "deployment apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) DeploymentReadyWait(
	ctx context.Context,
	log types.LoggerFactory,
	name, namespace string) error {

	expoBackoff := helpers.NewBackOff(
		5*time.Second,
		2,
		int(consts.CounterMaxRetryCount),
	)
	var (
		status *appsv1.Deployment
	)
	_err := expoBackoff.Run(
		ctx,
		log,
		func() (err error) {
			status, err = k.clientset.
				AppsV1().
				Deployments(namespace).
				Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "failed to get", "Reason", err))
			}
			return nil
		},
		func() bool {
			return status.Status.ReadyReplicas > 0
		},
		func(err error) (errW error, escalateErr bool) {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "deployment get failed", "Reason", err),
			), true
		},
		func() error {
			log.Success(ctx, "Few of the replica are ready", "readyReplicas", status.Status.ReadyReplicas)
			return nil
		},
		fmt.Sprintf("retrying no of ready replicas == 0 %s", name),
	)
	if _err != nil {
		return _err
	}

	return nil
}

func (k *K8sClient) DaemonsetDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.DaemonSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		DaemonSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "daemonset delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) DeploymentDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.Deployment) error {

	ns := o.Namespace
	err := k.clientset.
		AppsV1().
		Deployments(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "deployment delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *K8sClient) StatefulSetApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.StatefulSet) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "statefulset apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "statefulset apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) StatefulSetDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *appsv1.StatefulSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		StatefulSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "statefulset delete failed", "Reason", err),
		)
	}
	return nil
}
