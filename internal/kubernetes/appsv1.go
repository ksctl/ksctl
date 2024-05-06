package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
)

func (k *Kubernetes) daemonsetApply(o *appsv1.DaemonSet) error {
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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) deploymentApply(o *appsv1.Deployment) error {
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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) deploymentReadyWait(name, namespace string) error {

	count := consts.KsctlCounterConsts(0)
	for {

		status, err := k.clientset.
			AppsV1().
			Deployments(namespace).
			Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return log.NewError(err.Error())
		}
		if status.Status.ReadyReplicas > 0 {
			log.Success(fmt.Sprintf("~~> Few of the replica are ready [%v]", status.Status.ReadyReplicas))
			break
		}
		count++
		if count == consts.CounterMaxRetryCount*2 {
			return log.NewError("max retry reached")
		}
		log.Warn(fmt.Sprintf("retrying current no of success [readyReplicas: %v]", status.Status.ReadyReplicas))
		time.Sleep(10 * time.Second)
	}
	return nil
}

func (k *Kubernetes) daemonsetDelete(o *appsv1.DaemonSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		DaemonSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (k *Kubernetes) deploymentDelete(o *appsv1.Deployment) error {
	ns := o.Namespace
	err := k.clientset.
		AppsV1().
		Deployments(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (k *Kubernetes) statefulSetApply(o *appsv1.StatefulSet) error {
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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) statefulSetDelete(o *appsv1.StatefulSet) error {
	ns := o.Namespace

	err := k.clientset.
		AppsV1().
		StatefulSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
