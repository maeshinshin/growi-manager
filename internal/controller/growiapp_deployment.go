package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

func (r *GrowiReconciler) reconcileGrowiappDeployment(ctx context.Context, growi *growiv1.Growi) error {
	var err error
	logger := logf.FromContext(ctx)
	growiappDeploymentName := getGrowiDeploymentName(*growi)
	growiappSecretName := getGrowiappSecretName(*growi)
	growiappLabels := getLabels(*growi, COMPONENT_GROWIAPP)

	// Check if the Growi deployment already exists
	currGrowiDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      growiappDeploymentName,
		Namespace: growi.Namespace,
	}, currGrowiDeployment)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get Growi deployment")
		return err
	}

	// Create the Growi deployment
	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	growiDeployment := appsv1apply.Deployment(
		growiappDeploymentName,
		growi.Namespace,
	).
		WithLabels(getLabels(*growi, COMPONENT_GROWIAPP)).
		WithOwnerReferences(ownerRef).
		WithSpec(
			appsv1apply.DeploymentSpec().
				WithReplicas(growi.Spec.GrowiAppSpec.Replicas).
				WithSelector(
					metav1apply.LabelSelector().
						WithMatchLabels(growiappLabels),
				).
				WithTemplate(
					corev1apply.PodTemplateSpec().
						WithLabels(growiappLabels).
						WithSpec(
							corev1apply.PodSpec().
								WithContainers(
									corev1apply.Container().
										WithName("growiapp").
										WithImage(getGrowiAppImage(*growi)).
										WithImagePullPolicy(corev1.PullIfNotPresent).
										WithPorts(
											corev1apply.ContainerPort().
												WithName("growiapp").
												WithContainerPort(3000).
												WithProtocol(corev1.ProtocolTCP),
										).
										WithEnvFrom(
											corev1apply.EnvFromSource().
												WithSecretRef(
													corev1apply.SecretEnvSource().
														WithName(growiappSecretName),
												),
										).
										WithEnv(
											corev1apply.EnvVar().
												WithName("MONGO_URI").
												WithValue(
													getMongodbURI(*growi, r.getMongodbUsername(ctx, growi), r.getMongodbPassword(ctx, growi)),
												),
											corev1apply.EnvVar().
												WithName("ELASTICSEARCH_URI").
												WithValue(getElasticsearchURI(*growi)),
										).
										WithLivenessProbe(
											corev1apply.Probe().
												WithInitialDelaySeconds(30).
												WithPeriodSeconds(10).
												WithFailureThreshold(5).
												WithHTTPGet(
													corev1apply.HTTPGetAction().
														WithPath("/_api/v3/healthcheck").
														WithPort(intstr.FromString("growiapp")),
												),
										),
								),
						),
				),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(growiDeployment)
	if err != nil {
		logger.Error(err, "unable to convert Growi deployment to unstructured")
		return err
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(currGrowiDeployment, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract deployment")
		return err
	}

	if equality.Semantic.DeepEqual(currApplyConfig, growiDeployment) {
		logger.Info("Growi deployment already exists, skipping creation")
		return nil
	}

	if err := r.updateGrowiAppStatus(ctx, growi, growiv1.StartingGrowiApp); err != nil {
		return err
	}

	logger.Info("Creating or updating Growi deployment", "name", growiappDeploymentName)
	if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to patch Growi deployment")
		return err
	}

	return nil
}
