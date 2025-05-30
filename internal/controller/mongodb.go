package controller

import (
	"context"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileMongodb(ctx context.Context, growi *growiv1.Growi) error {
	// reconcile MongoDB secret
	if err := r.reconcileMongodbSecret(ctx, growi); err != nil {
		return err
	}

	// reconcile MongoDB headless service
	if err := r.reconcileMongodbHeadlessService(ctx, growi); err != nil {
		return err
	}

	// reconcile MongoDB service
	if err := r.reconcileMongodbService(ctx, growi); err != nil {
		return err
	}

	// reconcile MongoDB statefulset
	if err := r.reconcileMongodbStatefulSet(ctx, growi); err != nil {
		return err
	}
	return nil
}
