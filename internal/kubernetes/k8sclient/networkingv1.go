package k8sclient

import (
	"context"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) NetPolicyApply(
	ctx context.Context,
	log types.LoggerFactory,
	o *networkingv1.NetworkPolicy) error {

	ns := o.Namespace
	_, err := k.clientset.
		NetworkingV1().
		NetworkPolicies(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				NetworkingV1().
				NetworkPolicies(ns).
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(ctx, "netpol apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(ctx, "netpol apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *K8sClient) NetPolicyDelete(
	ctx context.Context,
	log types.LoggerFactory,
	o *networkingv1.NetworkPolicy) error {

	ns := o.Namespace
	err := k.clientset.
		NetworkingV1().
		NetworkPolicies(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(ctx, "netpol delete failed", "Reason", err),
		)
	}
	return nil
}
