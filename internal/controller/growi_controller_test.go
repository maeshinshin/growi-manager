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
	"time" // timeパッケージをインポート

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/maeshinshin/growi-manager/api/v1"
)

const (
	timeout  = time.Second * 10       // 10秒待機
	interval = time.Millisecond * 250 // 250ミリ秒ごとにチェック
)

var _ = Describe("Growi Controller", func() {
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		namespaceTypeNamespacedName := types.NamespacedName{
			Name: testNamespaceName,
		}

		growiTypeNamespacedName := types.NamespacedName{
			Name:      testGrowiName,
			Namespace: testNamespaceName,
		}

		mongodbSecretTypeNamespcedName := types.NamespacedName{
			Name: getMongodbSecretName(
				appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name: testGrowiName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		mongodbStatefulSetTypeNamespacedName := types.NamespacedName{
			Name: getMongodbStatefulSetName(
				appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name: testGrowiName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		BeforeEach(func() {
			var err error
			By("Existing the test namespace")
			namespace := &corev1.Namespace{}
			err = k8sClient.Get(ctx, namespaceTypeNamespacedName, namespace)
			Expect(err).NotTo(HaveOccurred())

			By("creating the custom resource for the Kind Growi")
			err = k8sClient.Get(ctx, growiTypeNamespacedName, testGrowi)
			if err != nil && apierrors.IsNotFound(err) {
				testGrowi = &appv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
					Spec: appv1.GrowiSpec{
						GrowiAppSpec: appv1.GrowiAppSpec{
							Version:  "7.2.2",
							Replicas: 1,
						},
						MongodbSpec: appv1.MongodbSpec{
							Version:  "6.0",
							Replicas: 3,
						},
						ElasticsearchSpec: appv1.ElasticsearchSpec{
							Version:  "8.7.0",
							Replicas: 3,
						},
						StorageClass: "standard",
					},
				}
				Expect(k8sClient.Create(ctx, testGrowi)).To(Succeed())
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

			Expect(err).NotTo(HaveOccurred())

			By("MongoDBSecret should be created")

			Eventually(func() error {
				err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, testMongodbSecret)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbSecret.Name).To(Equal(testMongodbSecretName))
			Expect(testMongodbSecret.Namespace).To(Equal(testNamespaceName))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testMongodbSecret.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbSecret.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[0]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[1]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[2]))

			By("MongodbStatefulSet should be created")
			Eventually(func() error {
				err = k8sClient.Get(ctx, mongodbStatefulSetTypeNamespacedName, testMongodbStatefulSet)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbStatefulSet.Name).To(Equal(testMongodbStatefulSetName))
			Expect(testMongodbStatefulSet.Namespace).To(Equal(testNamespaceName))

			By("Delete and recreate the MongoDBSecret")
			Expect(k8sClient.Delete(ctx, testMongodbSecret)).To(Succeed())

			Eventually(func() error {
				err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, testMongodbSecret)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbSecret.Name).To(Equal(testMongodbSecretName))
			Expect(testMongodbSecret.Namespace).To(Equal(testNamespaceName))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testMongodbSecret.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbSecret.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[0]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[1]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[2]))

			By("Deleting the custom resource")
			Expect(k8sClient.Delete(ctx, testGrowi)).To(Succeed())

			err = k8sClient.Get(ctx, growiTypeNamespacedName, testGrowi)
			Eventually(func() error {
				testGrowi = &appv1.Growi{}
				err = k8sClient.Get(ctx, growiTypeNamespacedName, testGrowi)
				if apierrors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("Growi resource is not deleted")
			}).Should(Succeed())

			By("MongoDBSecret should not be deleted")
			oldMongoDBSecret := testMongodbSecret.DeepCopy()
			err = k8sClient.Get(ctx, mongodbSecretTypeNamespcedName, testMongodbSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(reflect.DeepEqual(oldMongoDBSecret, testMongodbSecret)).To(BeTrue())
		})
	})
})
