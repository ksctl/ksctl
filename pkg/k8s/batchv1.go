package k8s

import (
	"context"

	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Client) JobApply(o *batchv1.Job) error {
	ns := o.Namespace

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
				return ksctlErrors.WrapError(
					ksctlErrors.ErrFailedKubernetesClient,
					k.l.NewError(k.ctx, "job apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKubernetesClient,
				k.l.NewError(k.ctx, "job apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Client) JobDelete(o *batchv1.Job) error {
	ns := o.Namespace

	err := k.clientset.
		BatchV1().
		Jobs(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKubernetesClient,
			k.l.NewError(k.ctx, "job delete failed", "Reason", err),
		)
	}
	return nil
}
