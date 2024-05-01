package kubernetes

import (
	"context"

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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
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
		return log.NewError(err.Error())
	}
	return nil
}

func (k *Kubernetes) clusterRoleBindingDelete(o *rbacv1.ClusterRoleBinding) error {

	err := k.clientset.
		RbacV1().
		ClusterRoleBindings().
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) roleDelete(o *rbacv1.Role, ns string) error {

	err := k.clientset.
		RbacV1().
		Roles(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (k *Kubernetes) roleApply(o *rbacv1.Role, ns string) error {

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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) roleBindingApply(o *rbacv1.RoleBinding, ns string) error {

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
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (k *Kubernetes) roleBindingDelete(o *rbacv1.RoleBinding, ns string) error {

	err := k.clientset.
		RbacV1().
		RoleBindings(ns).
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}
