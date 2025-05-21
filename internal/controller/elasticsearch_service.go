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

func (r *GrowiReconciler) reconcileElasticsearchHeadlessService(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	elasticsearchHeadlessServiceName := getElasticsearchHeadlessServiceName(*growi)
	elasticsearchLabels := getLabels(*growi, COMPONENT_ELASTICSEARCH)

	currElasticsearchHeadlessService := corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: elasticsearchHeadlessServiceName, Namespace: growi.Namespace}, &currElasticsearchHeadlessService)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get elasticsearch service")
		return err
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	elasticsearchHeadlessService := corev1apply.Service(elasticsearchHeadlessServiceName, growi.Namespace).
		WithLabels(elasticsearchLabels).
		WithOwnerReferences(ownerRef).
		WithSpec(
			corev1apply.ServiceSpec().
				WithClusterIP(corev1.ClusterIPNone).
				WithPorts(
					corev1apply.ServicePort().
						WithName("http").
						WithPort(9200).
						WithProtocol(corev1.ProtocolTCP),
					corev1apply.ServicePort().
						WithName("transport").
						WithPort(9300).
						WithProtocol(corev1.ProtocolTCP),
				).
				WithPublishNotReadyAddresses(true).
				WithSelector(elasticsearchLabels),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(elasticsearchHeadlessService)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(&currElasticsearchHeadlessService, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract service")
		return err
	}

	// Check the difference between the current and desired state
	if equality.Semantic.DeepEqual(currApplyConfig, elasticsearchHeadlessService) {
		logger.Info("Elasticsearch headless service already exists, skipping creation")
		return nil
	}

	logger.Info("Creating elasticsearch headless service")
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create elasticsearch headless service")
		return err
	}

	return nil
}

func (r GrowiReconciler) reconcileElasticsearchService(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	elasticsearchServiceName := getElasticsearchServiceName(*growi)
	elasticsearchLabels := getLabels(*growi, COMPONENT_ELASTICSEARCH)

	currElasticsearchService := corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: elasticsearchServiceName, Namespace: growi.Namespace}, &currElasticsearchService)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get elasticsearch secret")
		return err
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	elasticsearchService := corev1apply.Service(elasticsearchServiceName, growi.Namespace).
		WithLabels(elasticsearchLabels).
		WithOwnerReferences(ownerRef).
		WithSpec(
			corev1apply.ServiceSpec().
				WithType(corev1.ServiceTypeClusterIP).
				WithPorts(
					corev1apply.ServicePort().
						WithName("http").
						WithPort(9200).
						WithProtocol(corev1.ProtocolTCP),
				).
				WithSelector(elasticsearchLabels),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(elasticsearchService)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(&currElasticsearchService, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract service")
		return err
	}

	// Check the difference between the current and desired state
	if equality.Semantic.DeepEqual(currApplyConfig, elasticsearchService) {
		logger.Info("Elasticsearch service already exists, skipping creation")
		return nil
	}

	logger.Info("Creating elasticsearch service")
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create elasticsearch service")
		return err
	}

	return nil
}
