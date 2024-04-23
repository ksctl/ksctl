package kubernetes

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) netPolicyApply(o *networkingv1.NetworkPolicy, ns string) error {

	_, err := this.clientset.
		NetworkingV1().
		NetworkPolicies(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
				NetworkingV1().
				NetworkPolicies(ns).
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

func (this *Kubernetes) netPolicyDelete(o *networkingv1.NetworkPolicy, ns string) error {

	err := this.clientset.
		NetworkingV1().
		NetworkPolicies(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
