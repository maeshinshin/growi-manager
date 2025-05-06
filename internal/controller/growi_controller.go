/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

// GrowiReconciler reconciles a Growi object
type GrowiReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.maeshinshin.github.io,resources=growis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.maeshinshin.github.io,resources=growis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.maeshinshin.github.io,resources=growis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Growi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *GrowiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// Fetch the Growi instance
	var growi growiv1.Growi
	if err := client.IgnoreNotFound(r.Get(ctx, req.NamespacedName, &growi)); err != nil {
		logger.Error(err, "unable to fetch Growi")
		return ctrl.Result{}, err
	}

	// Update status if not set
	if growi.Status.MongoDBSecretStatus == nil {
		growi.Status.MongoDBSecretStatus = ptr.To(growiv1.WaitingOtherProcessMongoDBSecret)
	}
	if growi.Status.GrowiAppStatus == nil {
		growi.Status.GrowiAppStatus = ptr.To(growiv1.WaitingOtherProcessGrowiApp)
	}
	if growi.Status.MongoDBStatus == nil {
		growi.Status.MongoDBStatus = ptr.To(growiv1.WaitingOtherProcessMongoDB)
	}
	if growi.Status.ElasticSearchStatus == nil {
		growi.Status.ElasticSearchStatus = ptr.To(growiv1.WaitingOtherProcessElasticSearch)
	}

	// Reconcile MongoDB secret
	if err := r.reconcileMongoDBSecret(ctx, growi); err != nil {
		logger.Error(err, "unable to reconcile MongoDB secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrowiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&growiv1.Growi{}).WithEventFilter(&predicate.GenerationChangedPredicate{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.Deployment{}).
		Named("growi").
		Complete(r)
}
