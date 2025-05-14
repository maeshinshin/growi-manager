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
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1 "github.com/maeshinshin/growi-manager/api/v1"
)

var _ = Describe("Growi Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		growiTypeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: testNamespaceName,
		}
		mongodbSecretTypeNamespcedName := types.NamespacedName{
			Name: getMongodbSecretName(
				appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}
		growi := &appv1.Growi{}
		mongodbSecret := &corev1.Secret{}

		BeforeEach(func() {
			var err error
			By("Existing the test namespace")
			namespace := &corev1.Namespace{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: testNamespaceName}, namespace)
			Expect(err).NotTo(HaveOccurred())

			By("creating the custom resource for the Kind Growi")
			err = k8sClient.Get(ctx, growiTypeNamespacedName, growi)
			if err != nil && apierrors.IsNotFound(err) {
				resource := &appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: testNamespaceName,
					},
					Spec: appv1.GrowiSpec{
						GrowiAppSpec: appv1.GrowiAppSpec{
							Version:  "7.2.2",
							Replicas: 1,
						},
						MongodbSpec: appv1.MongodbSpec{
							Version:  "6.0",
							Replicas: 1,
						},
						ElasticsearchSpec: appv1.ElasticsearchSpec{
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
			By("Ensure Growi resource is deleted")
			growi := &appv1.Growi{}
			err := k8sClient.DeleteAllOf(ctx, growi, client.InNamespace(testNamespaceName))
			Expect(err == nil || apierrors.IsNotFound(err)).To(BeTrue())
			Eventually(func() error {
				growi := &appv1.GrowiList{}
				err = k8sClient.List(ctx, growi, client.InNamespace(testNamespaceName))
				Expect(err).NotTo(HaveOccurred())
				if len(growi.Items) == 0 {
					return nil
				}
				return fmt.Errorf("Growi resource is not deleted")
			}).Should(Succeed())

			By("Ensure Secret resource is deleted")
			secret := &corev1.Secret{}
			err = k8sClient.DeleteAllOf(ctx, secret, client.InNamespace(testNamespaceName))
			Expect(err == nil || apierrors.IsNotFound(err)).To(BeTrue())
			Eventually(func() error {
				secret := &corev1.SecretList{}
				err = k8sClient.List(ctx, secret, client.InNamespace(testNamespaceName))
				Expect(err).NotTo(HaveOccurred())
				if len(secret.Items) == 0 {
					return nil
				}
				return fmt.Errorf("Secret resource is not deleted")
			}).Should(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &GrowiReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: growiTypeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("MongoDBSecret should be created")
			mongodbSecret = &corev1.Secret{}
			err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, mongodbSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(mongodbSecret.Name).To(Equal("test-resource-mongodb-secret"))
			Expect(mongodbSecret.Namespace).To(Equal(testNamespaceName))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", resourceName))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(mongodbSecret.GetManagedFields()).To(HaveLen(1))
			Expect(mongodbSecret.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(mongodbSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_USERNAME"))
			Expect(mongodbSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_PASSWORD"))

			By("Delete and recreate the MongoDBSecret")
			Expect(k8sClient.Delete(ctx, mongodbSecret)).To(Succeed())
			mongodbSecret = &corev1.Secret{}
			err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, mongodbSecret)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: growiTypeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				mongodbSecret = &corev1.Secret{}
				err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, mongodbSecret)
				if err != nil {
					return err
				}
				return nil
			}).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(mongodbSecret.Name).To(Equal("test-resource-mongodb-secret"))
			Expect(mongodbSecret.Namespace).To(Equal(testNamespaceName))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", resourceName))
			Expect(mongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(mongodbSecret.GetManagedFields()).To(HaveLen(1))
			Expect(mongodbSecret.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(mongodbSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_USERNAME"))
			Expect(mongodbSecret.Data).To(HaveKey("MONGO_INITDB_ROOT_PASSWORD"))

			By("Deleting the custom resource")
			Expect(k8sClient.Delete(ctx, growi)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: growiTypeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.Get(ctx, growiTypeNamespacedName, growi)
			Eventually(func() error {
				growi = &appv1.Growi{}
				err = k8sClient.Get(ctx, growiTypeNamespacedName, growi)
				if apierrors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("Growi resource is not deleted")
			}).Should(Succeed())

			By("MongoDBSecret should not be deleted")
			oldMongoDBSecret := mongodbSecret.DeepCopy()
			err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, mongodbSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(reflect.DeepEqual(oldMongoDBSecret, mongodbSecret)).To(BeTrue())
		})
	})
})
