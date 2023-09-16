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
				return err
			}

		} else {
			return err
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
				return err
			}

		} else {
			return err
		}
	}
	return nil
}
