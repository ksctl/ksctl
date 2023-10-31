package universal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) configMapApply(o *corev1.ConfigMap, ns string) error {

	_, err := this.clientset.CoreV1().ConfigMaps(ns).Create(context.Background(), o, v1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.CoreV1().ConfigMaps(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) serviceApply(o *corev1.Service, ns string) error {

	_, err := this.clientset.CoreV1().Services(ns).Create(context.Background(), o, v1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.CoreV1().Services(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) namespaceCreate(ns string) error {

	if _, err := this.clientset.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: ns}}, v1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.CoreV1().Namespaces().Update(context.Background(), &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: ns}}, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) secretApply(o *corev1.Secret, ns string) error {

	_, err := this.clientset.CoreV1().Secrets(ns).Create(context.Background(), o, v1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.CoreV1().Secrets(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) PodApply(o *corev1.Pod, ns string) error {

	_, err := this.clientset.CoreV1().Pods(ns).Create(context.Background(), o, v1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.CoreV1().Pods(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) serviceaccountApply(o *corev1.ServiceAccount, ns string) error {

	_, err := this.clientset.CoreV1().ServiceAccounts(ns).Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.CoreV1().ServiceAccounts(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) nodesList() (*corev1.NodeList, error) {
	return this.clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
}

func (this *Kubernetes) nodeDelete(nodeName string) error {
	return this.clientset.CoreV1().Nodes().Delete(context.Background(), nodeName, v1.DeleteOptions{})
}
