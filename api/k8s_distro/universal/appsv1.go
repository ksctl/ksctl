package universal

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
)

func (this *Kubernetes) deploymentApply(o *appsv1.Deployment, ns string) error {

	_, err := this.clientset.AppsV1().Deployments(ns).Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.AppsV1().Deployments(ns).Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
