package controller

import (
	"context"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileGrowiapp(ctx context.Context, growi *growiv1.Growi) error {
	// reconcile Growiapp secret
	if err := r.reconcileGrowiappSecret(ctx, growi); err != nil {
		return err
	}

	// reconcile Growiapp service
	if err := r.reconcileGrowiappService(ctx, growi); err != nil {
		return err
	}

	// reconcile Growiapp deployment
	if err := r.reconcileGrowiappDeployment(ctx, growi); err != nil {
		return err
	}
	return nil
}
