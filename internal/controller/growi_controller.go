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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	logger.Info("Reconciling Growi", "name", req.Name, "namespace", req.Namespace)

	// Fetch the Growi instance
	var growi growiv1.Growi
	if err := r.Get(ctx, req.NamespacedName, &growi); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Growi resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "unable to fetch Growi")
			return ctrl.Result{}, err
		}
	}

	// Check if the Growi instance is marked for deletion
	if !growi.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Growi is being deleted")
		r.deleteFinalizer(ctx, &growi)
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if err := r.addFinalizer(ctx, &growi); err != nil {
		logger.Error(err, "unable to add finalizer")
		return ctrl.Result{}, err
	}

	// Update status if not set
	old := growi.DeepCopy()
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

	// update status
	if reflect.DeepEqual(old.Status, growi.Status) {
		if err := r.Status().Update(ctx, &growi); err != nil {
			logger.Error(err, "unable to update Growi status")
			return ctrl.Result{}, err
		}
	}

	// Reconcile MongoDB secret
	if err := r.reconcileMongoDBSecret(ctx, &growi); err != nil {
		logger.Error(err, "unable to reconcile MongoDB secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) addFinalizer(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	if !controllerutil.ContainsFinalizer(growi, FINALIZER_NAME) {
		logger.Info("Adding finalizer for the Growi")
		controllerutil.AddFinalizer(growi, FINALIZER_NAME)
	}
	if err := r.Update(ctx, growi); err != nil {
		logger.Error(err, "unable to update Growi with finalizer")
		return err
	}
	return nil
}

func (r *GrowiReconciler) deleteFinalizer(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	if controllerutil.ContainsFinalizer(growi, FINALIZER_NAME) {
		logger.Info("Removing finalizer for the Growi")
		controllerutil.RemoveFinalizer(growi, FINALIZER_NAME)
	}
	if err := r.Update(ctx, growi); err != nil {
		logger.Error(err, "unable to update Growi with finalizer")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrowiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&growiv1.Growi{},
			builder.WithPredicates(
				predicate.Or(
					predicate.Not(
						predicate.ResourceVersionChangedPredicate{},
					),
					predicate.GenerationChangedPredicate{},
				),
			),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []ctrl.Request {
					secret, ok := obj.(*corev1.Secret)
					if !ok {
						return nil
					}
					if secret.Labels["app.kubernetes.io/name"] == "growi" &&
						secret.Labels["app.kubernetes.io/managed-by"] == "growi-manager" &&
						secret.Labels["app.kubernetes.io/instance"] != "" {
						return []ctrl.Request{
							{
								NamespacedName: client.ObjectKey{
									Name:      secret.Labels["app.kubernetes.io/instance"],
									Namespace: secret.Namespace,
								},
							},
						}
					}
					return nil
				},
			),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool { return false },
			}),
		).
		Owns(&appsv1.Deployment{}).
		Named("growi").
		Complete(r)
}
