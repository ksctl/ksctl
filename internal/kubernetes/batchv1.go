package kubernetes

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) jobApply(o *batchv1.Job, ns string) error {

	_, err := this.clientset.
		BatchV1().
		Jobs(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
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
func (this *Kubernetes) jobDelete(o *batchv1.Job, ns string) error {

	err := this.clientset.
		BatchV1().
		Jobs(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
