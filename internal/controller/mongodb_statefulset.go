package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileMongodbStatefulSet(ctx context.Context, growi *growiv1.Growi) error {
	var err error
	var shouldCreateJob bool
	logger := logf.FromContext(ctx)
	mongodbStatefulSetName := getMongodbStatefulSetName(*growi)
	mongodbSecretName := getMongodbSecretName(*growi)
	mongodbHeadlessServiceName := getMongodbHeadlessServiceName(*growi)
	mongodbStatefulsetLabels := getLabels(*growi, COMPONENT_MONGODB)
	mongodbPersistentVolumeClaimName := getMongodbPersistentVolumeClaimName(*growi)
	mongodbPersistentVolumeClaimNameOfZero := getMongodbPersistentVolumeClaimNameOfZero(*growi)

	// Check if the MongoDB persistent volume claim already exists
	currMongodbPVC := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      mongodbPersistentVolumeClaimNameOfZero,
		Namespace: growi.Namespace,
	}, currMongodbPVC)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get MongoDB persistent volume claim")
	} else if err != nil && apierrors.IsNotFound(err) {
		shouldCreateJob = true
	}

	// Check if the MongoDB statefulset already exists
	currMongodbStatefulSet := &appsv1.StatefulSet{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      mongodbStatefulSetName,
		Namespace: growi.Namespace,
	}, currMongodbStatefulSet)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get MongoDB statefulset")
		return err
	}

	// Create the MongoDB statefulset
	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	mongoStatefulSet := appsv1apply.StatefulSet(
		mongodbStatefulSetName,
		growi.Namespace,
	).
		WithLabels(mongodbStatefulsetLabels).
		WithOwnerReferences(ownerRef).
		WithSpec(
			appsv1apply.StatefulSetSpec().
				WithServiceName(mongodbHeadlessServiceName).
				WithReplicas(growi.Spec.MongodbSpec.Replicas).
				WithPodManagementPolicy(appsv1.ParallelPodManagement).
				WithSelector(
					metav1apply.LabelSelector().
						WithMatchLabels(mongodbStatefulsetLabels),
				).
				WithTemplate(
					corev1apply.PodTemplateSpec().
						WithLabels(mongodbStatefulsetLabels).
						WithSpec(
							corev1apply.PodSpec().
								WithInitContainers(
									corev1apply.Container().
										WithName("init-mongo-key").
										WithImage(getMongodbImage(*growi)).
										WithCommand(
											"/bin/sh",
											"-c",
											"cp -L /etc/mongo-key/mongo.key /etc/mongo-key-copy/mongo.key && chown mongodb: /etc/mongo-key-copy/mongo.key && chmod 400 /etc/mongo-key-copy/mongo.key",
										).
										WithVolumeMounts(
											corev1apply.VolumeMount().
												WithName("mongo-key").
												WithMountPath("/etc/mongo-key"),
										).
										WithVolumeMounts(
											corev1apply.VolumeMount().
												WithName("mongo-key-copy").
												WithMountPath("/etc/mongo-key-copy"),
										).
										WithSecurityContext(
											corev1apply.SecurityContext().
												WithRunAsUser(0).
												WithRunAsGroup(0),
										),
								).
								WithContainers(
									corev1apply.Container().
										WithName("mongodb").
										WithImage(getMongodbImage(*growi)).
										WithPorts(corev1apply.ContainerPort().
											WithName("mongodb").
											WithContainerPort(27017),
										).
										WithArgs(
											"mongod",
											"--auth",
											"--replSet",
											mongodbStatefulSetName,
											"--bind_ip_all",
											"--keyFile",
											"/etc/mongo-key/mongo.key",
										).
										WithEnv(
											corev1apply.EnvVar().
												WithName("MONGO_REPLICA_SET_NAME").
												WithValue(mongodbStatefulSetName),
										).
										WithEnv(
											corev1apply.EnvVar().
												WithName("MONGO_INITDB_ROOT_USERNAME").
												WithValueFrom(corev1apply.EnvVarSource().
													WithSecretKeyRef(corev1apply.SecretKeySelector().
														WithName(mongodbSecretName).
														WithKey("MONGO_INITDB_ROOT_USERNAME"),
													),
												),
											corev1apply.EnvVar().
												WithName("MONGO_INITDB_ROOT_PASSWORD").
												WithValueFrom(corev1apply.EnvVarSource().
													WithSecretKeyRef(corev1apply.SecretKeySelector().
														WithName(mongodbSecretName).
														WithKey("MONGO_INITDB_ROOT_PASSWORD"),
													),
												),
										).
										WithVolumeMounts(
											corev1apply.VolumeMount().
												WithName(getMongodbPersistentVolumeClaimName(*growi)).
												WithMountPath("/data/db"),
											corev1apply.VolumeMount().
												WithName("mongo-key-copy").
												WithMountPath("/etc/mongo-key"),
										).
										WithLivenessProbe(
											corev1apply.Probe().
												WithTCPSocket(
													corev1apply.TCPSocketAction().
														WithPort(intstr.IntOrString{IntVal: 27017}),
												).
												WithInitialDelaySeconds(30).
												WithPeriodSeconds(10).
												WithTimeoutSeconds(5).
												WithFailureThreshold(5),
										).
										WithReadinessProbe(
											corev1apply.Probe().
												WithExec(
													corev1apply.ExecAction().
														WithCommand(
															"mongosh",
															"--eval",
															"--username $(MONGO_INITDB_ROOT_USERNAME)",
															"--password $(MONGO_INITDB_ROOT_PASSWORD)",
															"--authenticationDatabase admin",
															"'quit(rs.status().ok ? 0 : 1)'",
														),
												).
												WithInitialDelaySeconds(60).
												WithPeriodSeconds(10).
												WithTimeoutSeconds(5).
												WithFailureThreshold(5),
										),
								).
								WithVolumes(
									corev1apply.Volume().
										WithName("mongo-key").
										WithSecret(
											corev1apply.SecretVolumeSource().
												WithSecretName(mongodbSecretName).
												WithItems(
													corev1apply.KeyToPath().
														WithKey("mongo.key").
														WithPath("mongo.key").
														WithMode(0400),
												),
										),
									corev1apply.Volume().
										WithName("mongo-key-copy").
										WithEmptyDir(
											corev1apply.EmptyDirVolumeSource().
												WithMedium(""),
										),
								),
						),
				).
				WithVolumeClaimTemplates(
					corev1apply.PersistentVolumeClaim(
						mongodbPersistentVolumeClaimName,
						growi.Namespace,
					).
						WithLabels(mongodbStatefulsetLabels).
						WithKind("PersistentVolumeClaim").
						WithAPIVersion("v1").
						WithSpec(
							corev1apply.PersistentVolumeClaimSpec().
								WithAccessModes(corev1.ReadWriteOnce).
								WithResources(
									corev1apply.VolumeResourceRequirements().
										WithRequests(corev1.ResourceList{
											corev1.ResourceStorage: apiresource.MustParse("10Gi"),
										}),
								).
								WithVolumeMode(corev1.PersistentVolumeFilesystem).
								WithStorageClassName(growi.Spec.StorageClass),
						),
				),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongoStatefulSet)
	if err != nil {
		return err
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := appsv1apply.ExtractStatefulSet(currMongodbStatefulSet, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract statefulset")
		return err
	}

	if !shouldPatch(ctx, currApplyConfig, mongoStatefulSet) {
		logger.Info("MongoDB statefulset already exists, skipping creation")
		return nil
	}

	if err := r.updateMongodbStatus(ctx, growi, growiv1.StartingMongodb); err != nil {
		return err
	}

	logger.Info("Creating MongoDB statefulset")
	if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB statefulset")
		return err
	}

	if shouldCreateJob {
		if err := r.createMongoDBJob(ctx, growi); err != nil {
			return err
		}
	}

	return nil
}
