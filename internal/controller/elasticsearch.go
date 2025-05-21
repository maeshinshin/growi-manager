package controller

import (
	"context"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileElasticsearch(ctx context.Context, growi *growiv1.Growi) error {
	// reconcile Elasticsearch headless service
	if err := r.reconcileElasticsearchHeadlessService(ctx, growi); err != nil {
		return err
	}

	// reconcile Elasticsearch service
	if err := r.reconcileElasticsearchService(ctx, growi); err != nil {
		return err
	}

	// reconcile Elasticsearch statefulset
	if err := r.reconcileElasticsearchStatefulSet(ctx, growi); err != nil {
		return err
	}
	return nil
}
