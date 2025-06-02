package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r *GrowiReconciler) reconcileMongodbHeadlessService(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	mongodbHeadlessServiceName := getMongodbHeadlessServiceName(*growi)

	currMongodbHeadlessService := corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: mongodbHeadlessServiceName, Namespace: growi.Namespace}, &currMongodbHeadlessService)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get MongoDB service")
		return err
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	mongodbHeadlessService := corev1apply.Service(mongodbHeadlessServiceName, growi.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "growi",
			"app.kubernetes.io/instance":   growi.Name,
			"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
			"app.kubernetes.io/component":  "mongodb",
		}).
		WithOwnerReferences(ownerRef).
		WithSpec(
			corev1apply.ServiceSpec().
				WithClusterIP(corev1.ClusterIPNone).
				WithPorts(
					corev1apply.ServicePort().
						WithName("mongodb").
						WithPort(27017).
						WithProtocol(corev1.ProtocolTCP),
				).
				WithPublishNotReadyAddresses(true).
				WithSelector(map[string]string{
					"app.kubernetes.io/name":       "growi",
					"app.kubernetes.io/instance":   growi.Name,
					"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
					"app.kubernetes.io/component":  "mongodb",
				}),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongodbHeadlessService)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(&currMongodbHeadlessService, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract service")
		return err
	}

	// Check the difference between the current and desired state
	if equality.Semantic.DeepEqual(currApplyConfig, mongodbHeadlessService) {
		logger.Info("MongoDB headless service already exists, skipping creation")
		return nil
	}

	logger.Info("Creating MongoDB headless service")
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB headless service")
		return err
	}

	return nil
}

func (r *GrowiReconciler) reconcileMongodbService(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	mongodbServiceName := getMongodbServiceName(*growi)

	currMongodbService := corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: mongodbServiceName, Namespace: growi.Namespace}, &currMongodbService)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get MongoDB secret")
		return err
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	mongodbService := corev1apply.Service(mongodbServiceName, growi.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "growi",
			"app.kubernetes.io/instance":   growi.Name,
			"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
			"app.kubernetes.io/component":  "mongodb",
		}).
		WithOwnerReferences(ownerRef).
		WithSpec(
			corev1apply.ServiceSpec().
				WithType(corev1.ServiceTypeClusterIP).
				WithPorts(
					corev1apply.ServicePort().
						WithName("mongodb").
						WithPort(27017).
						WithProtocol(corev1.ProtocolTCP),
				).
				WithSelector(map[string]string{
					"app.kubernetes.io/name":       "growi",
					"app.kubernetes.io/instance":   growi.Name,
					"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
					"app.kubernetes.io/component":  "mongodb",
				}),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongodbService)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(&currMongodbService, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract service")
		return err
	}

	// Check the difference between the current and desired state
	if equality.Semantic.DeepEqual(currApplyConfig, mongodbService) {
		logger.Info("MongoDB service already exists, skipping creation")
		return nil
	}

	logger.Info("Creating or updating MongoDB service", "name", mongodbServiceName)
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB service")
		return err
	}

	return nil
}
