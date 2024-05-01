package kubernetes

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) jobApply(o *batchv1.Job, ns string) error {

	_, err := k.clientset.
		BatchV1().
		Jobs(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				BatchV1().
				Jobs(ns).
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}
func (k *Kubernetes) jobDelete(o *batchv1.Job, ns string) error {

	err := k.clientset.
		BatchV1().
		Jobs(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
