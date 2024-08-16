package k8sclient

import (
	"context"

	"github.com/ksctl/ksctl/pkg/types"
	nodev1 "k8s.io/api/node/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (k *K8sClient) RuntimeApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *nodev1.RuntimeClass,
) error {
	_, err := k.clientset.NodeV1().RuntimeClasses().Create(
		context.Background(),
		o,
		metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.NodeV1().RuntimeClasses().Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "runtime class apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "runtime class apply failed", "Reason", err),
			)
		}
	}

	return nil
}

func (k *K8sClient) RuntimeDelete(
	ctx context.Context,
	log types.LoggerFactory,
	resName string,
) error {
	err := k.clientset.NodeV1().RuntimeClasses().Delete(
		context.Background(),
		resName,
		metav1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "runtime class delete failed", "Reason", err),
		)
	}

	return nil
}
