/*
Copyright 2025.

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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1 "github.com/maeshinshin/growi-manager/api/v1"
)

var _ = Describe("Growi Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const namespaceName = "test-namespace"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespaceName,
		}
		growi := &appv1.Growi{}

		BeforeEach(func() {
			By("Creating the test namespace")
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).NotTo(HaveOccurred())

			By("Existing the test namespace")
			err = k8sClient.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace)
			Expect(err).NotTo(HaveOccurred())
			By("creating the custom resource for the Kind Growi")
			err = k8sClient.Get(ctx, typeNamespacedName, growi)
			if err != nil && errors.IsNotFound(err) {
				resource := &appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespaceName,
					},
					Spec: appv1.GrowiSpec{
						GrowiAppSpec: appv1.GrowiAppSpec{
							Version:  "7.2.2",
							Replicas: 1,
						},
						MongoDBSpec: appv1.MongoDBSpec{
							Version:  "6.0",
							Replicas: 1,
						},
						ElasticSearchSpec: appv1.ElasticSearchSpec{
							Version:  "8.7.0",
							Replicas: 1,
						},
						StorageClass: "standard",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &appv1.Growi{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Growi")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &GrowiReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Status MongoDBSecret status should be set to Exists")
			Eventually(func() error {
				growi = &appv1.Growi{}
				err = k8sClient.Get(ctx, typeNamespacedName, growi)
				if err != nil {
					return err
				}
				if *growi.Status.MongoDBSecretStatus == appv1.ExistMongoDBSecret {
					return nil
				}
				return fmt.Errorf("MongoDBSecret status is not set to Exists")
			}).Should(Succeed())

			By("MongoDBSecret should be created")
			mongoDBSecret := corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: getMongoDBSecretName(appv1.Growi{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}), Namespace: namespaceName}, &mongoDBSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(mongoDBSecret.Name).To(Equal("test-resource-mongodb-secret"))
			Expect(mongoDBSecret.Namespace).To(Equal(namespaceName))
			Expect(mongoDBSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_USERNAME"))
			Expect(mongoDBSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_PASSWORD"))
		})
	})
})
