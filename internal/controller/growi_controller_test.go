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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	growiappv1 "github.com/maeshinshin/growi-manager/api/v1"
)

var _ = Describe("Growi Controller", func() {
	const (
		timeout  = time.Second * 20
		interval = time.Millisecond * 250
	)
	var (
		testGrowi                        *growiappv1.Growi   = &growiappv1.Growi{}
		testMongodbHeadlessService       *corev1.Service     = &corev1.Service{}
		testMongodbService               *corev1.Service     = &corev1.Service{}
		testMongodbSecret                *corev1.Secret      = &corev1.Secret{}
		testMongodbStatefulSet           *appsv1.StatefulSet = &appsv1.StatefulSet{}
		testMongodbJob                   *batchv1.Job        = &batchv1.Job{}
		testElasticsearchHeadlessService *corev1.Service     = &corev1.Service{}
		testElasticsearchService         *corev1.Service     = &corev1.Service{}
		testElasticsearchStatefulSet     *appsv1.StatefulSet = &appsv1.StatefulSet{}
	)

	Context("When reconciling a resource", func() {
		ctx := context.Background()

		namespaceTypeNamespacedName := types.NamespacedName{
			Name: testNamespaceName,
		}

		growiTypeNamespacedName := types.NamespacedName{
			Name:      testGrowiName,
			Namespace: testNamespaceName,
		}

		mongodbHeadlessServiceTypeNamespacedName := types.NamespacedName{
			Name: getMongodbHeadlessServiceName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		mongodbServiceTypeNamespacedName := types.NamespacedName{
			Name: getMongodbServiceName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		mongodbSecretTypeNamespcedName := types.NamespacedName{
			Name: getMongodbSecretName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		mongodbStatefulSetTypeNamespacedName := types.NamespacedName{
			Name: getMongodbStatefulSetName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		mongodbJobTypeNamespacedName := types.NamespacedName{
			Name: getMongodbJobName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		elasticsearchHeadlessServiceTypeNamespacedName := types.NamespacedName{
			Name: getElasticsearchHeadlessServiceName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		elasticsearchServiceTypeNamespacedName := types.NamespacedName{
			Name: getElasticsearchServiceName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
				},
			),
			Namespace: testNamespaceName,
		}

		elasticsearchStatefulSetTypeNamespacedName := types.NamespacedName{
			Name: getElasticsearchStatefulSetName(
				growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
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
				testGrowi = &growiappv1.Growi{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testGrowiName,
						Namespace: testNamespaceName,
					},
					Spec: growiappv1.GrowiSpec{
						GrowiAppSpec: growiappv1.GrowiAppSpec{
							Version:  "7.2.2",
							Replicas: 1,
						},
						MongodbSpec: growiappv1.MongodbSpec{
							Version:  "6.0",
							Replicas: 3,
						},
						ElasticsearchSpec: growiappv1.ElasticsearchSpec{
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
			growi := &growiappv1.Growi{}
			err := k8sClient.DeleteAllOf(ctx, growi, client.InNamespace(testNamespaceName))
			Expect(err == nil || apierrors.IsNotFound(err)).To(BeTrue())
			Eventually(func() error {
				growi := &growiappv1.GrowiList{}
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
			By("MongodbHeadlessService should be created")
			Eventually(func() error {
				err := k8sClient.Get(ctx, mongodbHeadlessServiceTypeNamespacedName, testMongodbHeadlessService)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbHeadlessService.Name).To(Equal(getMongodbHeadlessServiceName(*testGrowi)))
			Expect(testMongodbHeadlessService.Namespace).To(Equal(testNamespaceName))
			Expect(testMongodbHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testMongodbHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
			Expect(testMongodbHeadlessService.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbHeadlessService.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testMongodbHeadlessService.Spec.ClusterIP).To(Equal(corev1.ClusterIPNone))
			Expect(testMongodbHeadlessService.Spec.Ports).To(HaveLen(1))
			Expect(testMongodbHeadlessService.Spec.Ports[0].Name).To(Equal("mongodb"))
			Expect(testMongodbHeadlessService.Spec.Ports[0].Port).To(Equal(int32(27017)))
			Expect(testMongodbHeadlessService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testMongodbHeadlessService.Spec.PublishNotReadyAddresses).To(BeTrue())
			Expect(testMongodbHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
			Expect(testMongodbHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))

			By("MongodbService should be created")
			Eventually(func() error {
				err = k8sClient.Get(ctx, mongodbServiceTypeNamespacedName, testMongodbService)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbService.Name).To(Equal(getMongodbServiceName(*testGrowi)))
			Expect(testMongodbService.Namespace).To(Equal(testNamespaceName))
			Expect(testMongodbService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testMongodbService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
			Expect(testMongodbService.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbService.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testMongodbService.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(testMongodbService.Spec.Ports).To(HaveLen(1))
			Expect(testMongodbService.Spec.Ports[0].Name).To(Equal("mongodb"))
			Expect(testMongodbService.Spec.Ports[0].Port).To(Equal(int32(27017)))
			Expect(testMongodbService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testMongodbService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
			Expect(testMongodbService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))

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
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
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

			By("MongodbJob should be created")
			Eventually(func() error {
				err = k8sClient.Get(ctx, mongodbJobTypeNamespacedName, testMongodbJob)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())
			Expect(err).NotTo(HaveOccurred())
			Expect(testMongodbJob.Name).To(Equal(getMongodbJobName(*testGrowi)))
			Expect(testMongodbJob.Namespace).To(Equal(testNamespaceName))
			Expect(testMongodbJob.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testMongodbJob.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testMongodbJob.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testMongodbJob.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "init-mongodb"))
			Expect(testMongodbJob.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbJob.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(*testMongodbJob.Spec.BackoffLimit).To(Equal(int32(5)))
			Expect(testMongodbJob.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyOnFailure))
			Expect(testMongodbJob.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Name).To(Equal("mongodb-init"))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Image).To(Equal(getMongodbImage(*testGrowi)))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"mongosh"}))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Args).To(Equal([]string{
				"--host",
				getMongodbPodOfZeroFQDN(*testGrowi),
				"--username",
				"$(MONGODB_USERNAME)",
				"--password",
				"$(MONGODB_PASSWORD)",
				"--eval",
				buildReplicaSetInitiateCommand(
					getMongodbStatefulSetName(*testGrowi),
					getMongodbHeadlessServiceName(*testGrowi),
					testNamespaceName,
					*testGrowi,
				),
			}))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env).To(HaveLen(2))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("MONGODB_USERNAME"))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.SecretKeyRef.Name).To(Equal(getMongodbSecretName(*testGrowi)))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[0].ValueFrom.SecretKeyRef.Key).To(Equal("MONGO_INITDB_ROOT_USERNAME"))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[1].Name).To(Equal("MONGODB_PASSWORD"))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[1].ValueFrom.SecretKeyRef.Name).To(Equal(getMongodbSecretName(*testGrowi)))
			Expect(testMongodbJob.Spec.Template.Spec.Containers[0].Env[1].ValueFrom.SecretKeyRef.Key).To(Equal("MONGO_INITDB_ROOT_PASSWORD"))
			Expect(testMongodbJob.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyOnFailure))
			Expect(*testMongodbJob.Spec.TTLSecondsAfterFinished).To(Equal(int32(60)))

			By("ElasticsearchHeadlessService should be created")
			Eventually(func() error {
				err := k8sClient.Get(ctx, elasticsearchHeadlessServiceTypeNamespacedName, testElasticsearchHeadlessService)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testElasticsearchHeadlessService.Name).To(Equal(testElasticsearchHeadlessServiceName))
			Expect(testElasticsearchHeadlessService.Namespace).To(Equal(testNamespaceName))
			Expect(testElasticsearchHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchHeadlessService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchHeadlessService.GetManagedFields()).To(HaveLen(1))
			Expect(testElasticsearchHeadlessService.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testElasticsearchHeadlessService.Spec.ClusterIP).To(Equal(corev1.ClusterIPNone))
			Expect(testElasticsearchHeadlessService.Spec.Ports).To(HaveLen(2))
			Expect(testElasticsearchHeadlessService.Spec.Ports[0].Name).To(Equal("http"))
			Expect(testElasticsearchHeadlessService.Spec.Ports[0].Port).To(Equal(int32(9200)))
			Expect(testElasticsearchHeadlessService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testElasticsearchHeadlessService.Spec.Ports[1].Name).To(Equal("transport"))
			Expect(testElasticsearchHeadlessService.Spec.Ports[1].Port).To(Equal(int32(9300)))
			Expect(testElasticsearchHeadlessService.Spec.Ports[1].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testElasticsearchHeadlessService.Spec.PublishNotReadyAddresses).To(BeTrue())
			Expect(testElasticsearchHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchHeadlessService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))

			By("ElasticsearchService should be created")
			Eventually(func() error {
				err = k8sClient.Get(ctx, elasticsearchServiceTypeNamespacedName, testElasticsearchService)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testElasticsearchService.Name).To(Equal(testElasticsearchServiceName))
			Expect(testElasticsearchService.Namespace).To(Equal(testNamespaceName))
			Expect(testElasticsearchService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchService.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchService.GetManagedFields()).To(HaveLen(1))
			Expect(testElasticsearchService.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testElasticsearchService.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(testElasticsearchService.Spec.Ports).To(HaveLen(1))
			Expect(testElasticsearchService.Spec.Ports[0].Name).To(Equal("http"))
			Expect(testElasticsearchService.Spec.Ports[0].Port).To(Equal(int32(9200)))
			Expect(testElasticsearchService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testElasticsearchService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchService.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))

			By("ElasticsearchStatefulSet should be created")
			Eventually(func() error {
				err = k8sClient.Get(ctx, elasticsearchStatefulSetTypeNamespacedName, testElasticsearchStatefulSet)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(err).NotTo(HaveOccurred())
			Expect(testElasticsearchStatefulSet.Name).To(Equal(getElasticsearchStatefulSetName(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Namespace).To(Equal(testNamespaceName))
			Expect(testElasticsearchStatefulSet.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchStatefulSet.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchStatefulSet.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchStatefulSet.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchStatefulSet.GetManagedFields()).To(HaveLen(1))
			Expect(testElasticsearchStatefulSet.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testElasticsearchStatefulSet.Spec.ServiceName).To(Equal(getElasticsearchHeadlessServiceName(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Replicas).To(Equal(&testGrowi.Spec.ElasticsearchSpec.Replicas))
			Expect(testElasticsearchStatefulSet.Spec.PodManagementPolicy).To(Equal(appsv1.ParallelPodManagement))
			Expect(testElasticsearchStatefulSet.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchStatefulSet.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchStatefulSet.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchStatefulSet.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchStatefulSet.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchStatefulSet.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchStatefulSet.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchStatefulSet.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers).To(HaveLen(2))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[0].Name).To(Equal("sysctl"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[0].Image).To(Equal("busybox"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[0].Command).To(Equal([]string{"sysctl", "-w", "vm.max_map_count=262144"}))
			Expect(*testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[0].SecurityContext.Privileged).To(BeTrue())
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].Name).To(Equal("install-plugins"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].Image).To(Equal(getElasticsearchImage(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].Command).To(Equal([]string{"/bin/sh", "-c", "rm /usr/share/elasticsearch/plugins/* -rf && bin/elasticsearch-plugin install analysis-icu && bin/elasticsearch-plugin install analysis-kuromoji"}))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(HaveLen(1))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts[0].Name).To(Equal("elasticsearch-plugin"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.InitContainers[1].VolumeMounts[0].MountPath).To(Equal("/usr/share/elasticsearch/plugins"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Name).To(Equal("elasticsearch"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Image).To(Equal(getElasticsearchImage(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports).To(HaveLen(2))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].Name).To(Equal("http"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(9200)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].Name).To(Equal("transport"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].ContainerPort).To(Equal(int32(9300)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Ports[1].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env).To(HaveLen(10))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("ES_JAVA_OPTS"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[0].Value).To(Equal("-Xms512m -Xmx512m"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[1].Name).To(Equal("cluster.name"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[1].Value).To(Equal(getElasticsearchStatefulSetName(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[2].Name).To(Equal("network.host"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[2].Value).To(Equal("0.0.0.0"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[3].Name).To(Equal("http.cors.enabled"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[3].Value).To(Equal("true"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[4].Name).To(Equal("http.cors.allow-origin"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[4].Value).To(Equal("\"*\""))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[5].Name).To(Equal("discovery.seed_hosts"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[5].Value).To(Equal(getElasticsearchHostList(testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[6].Name).To(Equal("cluster.initial_master_nodes"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[6].Value).To(Equal(getElasticsearchHostList(testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[7].Name).To(Equal("node.roles"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[7].Value).To(Equal("[master, data, ingest]"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[8].Name).To(Equal("bootstrap.memory_lock"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[8].Value).To(Equal("false"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[9].Name).To(Equal("xpack.security.enabled"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].Env[9].Value).To(Equal("false"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts).To(HaveLen(2))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("elasticsearch-plugin"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/usr/share/elasticsearch/plugins"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).To(Equal(getElasticsearchDataPersistentVolumeClaimName(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).To(Equal("/usr/share/elasticsearch/data"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).To(Equal("/"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.String()).To(Equal("http"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds).To(Equal(int32(60)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds).To(Equal(int32(5)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds).To(Equal(int32(10)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold).To(Equal(int32(1)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold).To(Equal(int32(5)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).To(Equal("/_cluster/health?local=true&wait_for_status=yellow&timeout=1s"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.String()).To(Equal("http"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds).To(Equal(int32(90)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds).To(Equal(int32(5)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds).To(Equal(int32(10)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.SuccessThreshold).To(Equal(int32(1)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.FailureThreshold).To(Equal(int32(5)))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.SecurityContext).ToNot(BeNil())
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Volumes).To(HaveLen(1))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Volumes[0].Name).To(Equal("elasticsearch-plugin"))
			Expect(testElasticsearchStatefulSet.Spec.Template.Spec.Volumes[0].EmptyDir).ToNot(BeNil())
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates).To(HaveLen(1))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Name).To(Equal(getElasticsearchDataPersistentVolumeClaimName(*testGrowi)))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Namespace).To(Equal(testNamespaceName))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", "growi"))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", testGrowiName))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", FIELDMANAGER_NAME))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "elasticsearch"))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.AccessModes).To(ContainElement(corev1.ReadWriteOnce))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests).To(HaveKeyWithValue(corev1.ResourceStorage, resource.MustParse("10Gi")))
			Expect(*testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.VolumeMode).To(Equal(corev1.PersistentVolumeFilesystem))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.StorageClassName).To(Equal(&testGrowi.Spec.StorageClass))
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.Selector).To(BeNil())
			Expect(testElasticsearchStatefulSet.Spec.VolumeClaimTemplates[0].Spec.DataSource).To(BeNil())

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
			Expect(testMongodbSecret.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "mongodb"))
			Expect(testMongodbSecret.GetManagedFields()).To(HaveLen(1))
			Expect(testMongodbSecret.GetManagedFields()[0].Manager).To(Equal(FIELDMANAGER_NAME))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[0]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[1]))
			Expect(testMongodbSecret.Data).To(HaveKey(testMongodbSecretKeyName[2]))

			By("Deleting the custom resource")
			Expect(k8sClient.Delete(ctx, testGrowi)).To(Succeed())

			err = k8sClient.Get(ctx, growiTypeNamespacedName, testGrowi)
			Eventually(func() error {
				testGrowi = &growiappv1.Growi{}
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
