package kubernetes

import (
	"context"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) clusterRoleApply(o *rbacv1.ClusterRole) error {

	_, err := k.clientset.
		RbacV1().
		ClusterRoles().
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				RbacV1().
				ClusterRoles().
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(kubernetesCtx, "clusterrole apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "clusterrole apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Kubernetes) clusterRoleDelete(o *rbacv1.ClusterRole) error {

	err := k.clientset.
		RbacV1().
		ClusterRoles().
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "clusterrole delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Kubernetes) clusterRoleBindingDelete(o *rbacv1.ClusterRoleBinding) error {

	err := k.clientset.
		RbacV1().
		ClusterRoleBindings().
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "clusterrolebinding delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Kubernetes) clusterRoleBindingApply(o *rbacv1.ClusterRoleBinding) error {

	_, err := k.clientset.
		RbacV1().
		ClusterRoleBindings().
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				RbacV1().
				ClusterRoleBindings().
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(kubernetesCtx, "clusterrolebinding apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "clusterrolebinding apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Kubernetes) roleDelete(o *rbacv1.Role) error {

	err := k.clientset.
		RbacV1().
		Roles(o.Namespace).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "role delete failed", "Reason", err),
		)
	}
	return nil
}

func (k *Kubernetes) roleApply(o *rbacv1.Role) error {

	ns := o.Namespace
	_, err := k.clientset.
		RbacV1().
		Roles(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				RbacV1().
				Roles(ns).
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(kubernetesCtx, "role apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "role apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Kubernetes) roleBindingApply(o *rbacv1.RoleBinding) error {
	ns := o.Namespace

	_, err := k.clientset.
		RbacV1().
		RoleBindings(ns).
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.clientset.
				RbacV1().
				RoleBindings(ns).
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(kubernetesCtx, "rolebinding apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "rolebinding apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Kubernetes) roleBindingDelete(o *rbacv1.RoleBinding) error {
	ns := o.Namespace

	err := k.clientset.
		RbacV1().
		RoleBindings(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "rolebinding delete failed", "Reason", err),
		)
	}
	return nil
}
