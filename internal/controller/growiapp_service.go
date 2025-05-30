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

func (r *GrowiReconciler) reconcileGrowiappService(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	growiappServiceName := getGrowiappServiceName(*growi)
	growiappLabels := getLabels(*growi, COMPONENT_GROWIAPP)

	currGrowiappService := corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: growiappServiceName, Namespace: growi.Namespace}, &currGrowiappService)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get Growiapp secret")
		return err
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	growiappService := corev1apply.Service(growiappServiceName, growi.Namespace).
		WithLabels(growiappLabels).
		WithOwnerReferences(ownerRef).
		WithSpec(
			corev1apply.ServiceSpec().
				WithType(corev1.ServiceTypeClusterIP).
				WithPorts(
					corev1apply.ServicePort().
						WithName("growiapp").
						WithPort(3000).
						WithProtocol(corev1.ProtocolTCP),
				).
				WithSelector(growiappLabels),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(growiappService)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(&currGrowiappService, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract service")
		return err
	}

	// Check the difference between the current and desired state
	if equality.Semantic.DeepEqual(currApplyConfig, growiappService) {
		logger.Info("Growiapp service already exists, skipping creation")
		return nil
	}

	logger.Info("Creating or updating Growiapp service", "name", growiappServiceName)
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create Growiapp service")
		return err
	}

	return nil
}
