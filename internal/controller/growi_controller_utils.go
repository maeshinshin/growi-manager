package controller

import (
	"context"

	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) updateMongodbStatus(ctx context.Context, growi *growiv1.Growi, status growiv1.MongodbStatusType) error {
	logger := logf.FromContext(ctx)

	growi.Status.MongodbStatus = ptr.To(status)
	if err := r.Status().Update(ctx, growi); err != nil {
		logger.Error(err, "Failed to update Growi status")
		return err
	}
	return nil
}

func (r GrowiReconciler) updateGrowiAppStatus(ctx context.Context, growi *growiv1.Growi, status growiv1.GrowiAppStatusType) error {
	logger := logf.FromContext(ctx)

	growi.Status.GrowiAppStatus = ptr.To(status)
	if err := r.Status().Update(ctx, growi); err != nil {
		logger.Error(err, "Failed to update Growi status")
		return err
	}
	return nil
}

func (r GrowiReconciler) updateElasticsearchStatus(ctx context.Context, growi *growiv1.Growi, status growiv1.ElasticsearchStatusType) error {
	logger := logf.FromContext(ctx)

	growi.Status.ElasticsearchStatus = ptr.To(status)
	if err := r.Status().Update(ctx, growi); err != nil {
		logger.Error(err, "Failed to update Growi status")
		return err
	}
	return nil
}

func (r GrowiReconciler) controllerReference(ctx context.Context, growi *growiv1.Growi) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	logger := logf.FromContext(ctx)

	gvk, err := apiutil.GVKForObject(growi, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to get GVK for object")
		return nil, err
	}

	return metav1apply.OwnerReference().
			WithAPIVersion(gvk.GroupVersion().String()).
			WithKind(gvk.Kind).
			WithName(growi.Name).
			WithUID(growi.UID).
			WithBlockOwnerDeletion(true).
			WithController(true),
		nil
}
