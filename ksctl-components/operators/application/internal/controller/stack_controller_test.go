/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	applicationv1alpha1 "github.com/ksctl/ksctl/ksctl-components/operators/application/api/v1alpha1"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
)

var _ = Describe("Stack Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.WithValue(
			context.Background(),
			consts.KsctlTestFlagKey,
			"true",
		)

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		stack := &applicationv1alpha1.Stack{}
		log = logger.NewStructuredLogger(
			LogVerbosity["DEBUG"],
			LogWriter)

		BeforeEach(func() {
			By("creating the custom resource for the Kind Stack")
			err := k8sClient.Get(ctx, typeNamespacedName, stack)
			if err != nil && errors.IsNotFound(err) {
				resource := &applicationv1alpha1.Stack{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: applicationv1alpha1.StackSpec{
						Stacks: []applicationv1alpha1.StackObj{
							{
								StackId: "demo-on-simple",
								AppType: applicationv1alpha1.TypeApp,
							},
							{
								StackId: "demo-on-helm-overrides",
								AppType: applicationv1alpha1.TypeApp,
								Overrides: &v1.JSON{
									Raw: []byte(`{"istio-component":{"version":"v0.0.1"},"abcd":{"version":"latest","helmValues":{"enableX": "true"}}}`),
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &applicationv1alpha1.Stack{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Stack")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &StackReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
