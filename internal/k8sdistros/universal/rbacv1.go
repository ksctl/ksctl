package universal

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) clusterroleApply(o *rbacv1.ClusterRole) error {

	_, err := this.clientset.RbacV1().ClusterRoles().Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.RbacV1().ClusterRoles().Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (this *Kubernetes) clusterrolebindingApply(o *rbacv1.ClusterRoleBinding) error {

	_, err := this.clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (this *Kubernetes) roleApply(o *rbacv1.Role, ns string) error {

	_, err := this.clientset.RbacV1().Roles(ns).Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.RbacV1().Roles(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (this *Kubernetes) rolebindingApply(o *rbacv1.RoleBinding, ns string) error {

	_, err := this.clientset.RbacV1().RoleBindings(ns).Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.RbacV1().RoleBindings(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
