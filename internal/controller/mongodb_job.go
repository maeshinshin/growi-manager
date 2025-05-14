package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	batchv1apply "k8s.io/client-go/applyconfigurations/batch/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r *GrowiReconciler) createMongoDBJob(ctx context.Context, growi *growiv1.Growi) error {
	var err error
	logger := logf.FromContext(ctx)
	jobName := getMongodbJobName(*growi)
	mongodbStatefulSetName := getMongodbStatefulSetName(*growi)
	mongodbSecretName := getMongodbSecretName(*growi)
	mongodbHeadlessServiceName := getMongodbHeadlessServiceName(*growi)

	// get the MongoDB job
	currMongodbJob := &batchv1.Job{}

	if err = r.Get(ctx, client.ObjectKey{
		Name:      jobName,
		Namespace: growi.Namespace,
	}, currMongodbJob); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get MongoDB job")
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	// create the MongoDB job
	mongodbJob := batchv1apply.Job(
		jobName, growi.Namespace).
		WithLabels(getLabels(*growi, COMPONENT_INIT_MONGODB)).
		WithOwnerReferences(ownerRef).
		WithSpec(batchv1apply.JobSpec().
			WithBackoffLimit(5).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(getLabels(*growi, COMPONENT_INIT_MONGODB)).
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName("mongodb-init").
						WithImage(getMongodbImage(*growi)).
						WithCommand("mongosh").
						WithArgs(
							"--host",
							getMongodbPodOfZeroFQDN(*growi),
							"--username",
							"$(MONGODB_USERNAME)",
							"--password",
							"$(MONGODB_PASSWORD)",
							"--eval",
							buildReplicaSetInitiateCommand(mongodbStatefulSetName, mongodbHeadlessServiceName, growi.Namespace, *growi),
						).
						WithEnv(
							corev1apply.EnvVar().
								WithName("MONGODB_USERNAME").
								WithValueFrom(corev1apply.EnvVarSource().
									WithSecretKeyRef(corev1apply.SecretKeySelector().
										WithName(mongodbSecretName).
										WithKey("MONGO_INITDB_ROOT_USERNAME"),
									),
								),
							corev1apply.EnvVar().
								WithName("MONGODB_PASSWORD").
								WithValueFrom(corev1apply.EnvVarSource().
									WithSecretKeyRef(corev1apply.SecretKeySelector().
										WithName(mongodbSecretName).
										WithKey("MONGO_INITDB_ROOT_PASSWORD"),
									),
								),
						),
					).
					WithRestartPolicy(corev1.RestartPolicyOnFailure),
				),
			).
			WithTTLSecondsAfterFinished(60),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongodbJob)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := batchv1apply.ExtractJob(currMongodbJob, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "Failed to extract MongoDB job")
		return err
	}

	if equality.Semantic.DeepEqual(currApplyConfig, mongodbJob) {
		logger.Info("MongoDB job already exists, skipping creation")
		return nil
	}

	logger.Info("Creating MongoDB job")
	if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB job")
		return err
	}

	return nil
}

func buildReplicaSetInitiateCommand(stsName, svcName, ns string, growi growiv1.Growi) string {
	members := ""
	replicas := int(growi.Spec.MongodbSpec.Replicas)
	for i := range replicas {
		// 各メンバーのホスト名は、StatefulSet Pod の安定したネットワーク識別子を使用
		host := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", stsName, i, svcName, ns)
		members += fmt.Sprintf("{ _id: %d, host: \"%s\" }", i, host)
		if i < replicas-1 {
			members += ", "
		}
	}

	command := fmt.Sprintf(`
		rs.initiate({
			_id: "%s",
			members: [ %s ]
		})
	`, stsName, members)

	return command
}
