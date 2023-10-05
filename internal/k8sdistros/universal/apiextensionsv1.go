package universal

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) apiextensionsApply(o *apiextensionsv1.CustomResourceDefinition) error {

	_, err := this.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), o, v1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Update(context.Background(), o, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
