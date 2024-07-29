package k8sclient

import (
	"context"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) JobApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *batchv1.Job) error {
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
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "job apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "job apply failed", "Reason", err),
			)
		}
	}
	return nil
}
func (k *K8sClient) JobDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *batchv1.Job) error {
	ns := o.Namespace

	err := k.clientset.
		BatchV1().
		Jobs(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "job delete failed", "Reason", err),
		)
	}
	return nil
}
