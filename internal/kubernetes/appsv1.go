package kubernetes

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
)

func (this *Kubernetes) deploymentApply(o *appsv1.Deployment, ns string) error {

	_, err := this.clientset.
		AppsV1().
		Deployments(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
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

func (this *Kubernetes) deploymentDelete(o *appsv1.Deployment, ns string) error {
	err := this.clientset.
		AppsV1().
		Deployments(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) statefulSetApply(o *appsv1.StatefulSet, ns string) error {

	_, err := this.clientset.
		AppsV1().
		StatefulSets(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
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

func (this *Kubernetes) statefulSetDelete(o *appsv1.StatefulSet, ns string) error {

	err := this.clientset.
		AppsV1().
		StatefulSets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
