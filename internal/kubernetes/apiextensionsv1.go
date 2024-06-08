package kubernetes

import (
	"context"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) apiExtensionsApply(o *apiextensionsv1.CustomResourceDefinition) error {

	_, err := k.apiextensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = k.apiextensionsClient.
				ApiextensionsV1().
				CustomResourceDefinitions().
				Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return ksctlErrors.ErrFailedKubernetesClient.Wrap(
					log.NewError(kubernetesCtx, "apiExtension apply failed", "Reason", err),
				)
			}
		} else {
			return ksctlErrors.ErrFailedKubernetesClient.Wrap(
				log.NewError(kubernetesCtx, "apiExtension apply failed", "Reason", err),
			)
		}
	}
	return nil
}

func (k *Kubernetes) apiExtensionsDelete(o *apiextensionsv1.CustomResourceDefinition) error {

	err := k.apiextensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Delete(context.Background(), o.Name, v1.DeleteOptions{})
	if err != nil {
		return ksctlErrors.ErrFailedKubernetesClient.Wrap(
			log.NewError(kubernetesCtx, "apiExtension delete failed", "Reason", err),
		)
	}
	return nil
}
