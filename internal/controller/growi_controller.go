/*
Copyright 2024.

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

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	maeshinshingithubiov1 "github.com/maeshinshin/growi-manager/api/v1"
)

// GrowiReconciler reconciles a Growi object
type GrowiReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=maeshinshin.github.io,resources=growis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maeshinshin.github.io,resources=growis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=maeshinshin.github.io,resources=growis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Growi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *GrowiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("--- Reconciliation triggered")

	var growi maeshinshingithubiov1.Growi
	if err := r.Get(ctx, req.NamespacedName, &growi); err != nil {
		logger.Error(err, "Fetch Growi object: failed")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Fetch Growi object: succeeded", "growi", growi.Name, "createdAt", growi.CreationTimestamp)

	if res, err := r.CreateOrUpdateGrowiDeployment(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateMongodbStatefulSet(ctx, req, growi, logger); err != nil {
		return res, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrowiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&maeshinshingithubiov1.Growi{}).
		Complete(r)
}

func (r *GrowiReconciler) CreateOrUpdateGrowiDeployment(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	// Create Deployment object if not exists
	growi_deployment := &appv1.Deployment{}
	growi_deployment.SetNamespace(growi.Spec.Growi_app_namespace)
	growi_deployment.SetName("growi-" + growi.Name + "-app")
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, growi_deployment, func() error {
		growi_deployment.Spec = appv1.DeploymentSpec{
			Replicas: pointer.Int32(growi.Spec.Growi_replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": growi_deployment.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": growi_deployment.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "growi",
							Image: "weseek/growi:" + growi.Spec.Growi_version,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Growi_deployment", op)
	}

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateGrowiService(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	growi_service := &corev1.Service{}
	growi_service.SetName("growi-" + growi.Name + "-app-service")
	growi_service.SetNamespace(growi.Spec.Growi_app_namespace)
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, growi_service, func() error {
		growi_service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": "growi-" + growi.Name + "-app"},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Growi_deployment", op)
	}
	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateMongodbStatefulSet(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	// Create Deployment object if not exists
	mongodb_statefulset := &appv1.StatefulSet{}
	mongodb_statefulset.SetNamespace(growi.Spec.Mongo_db_namespace)
	mongodb_statefulset.SetName("growi-" + growi.Name + "-db")
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, mongodb_statefulset, func() error {
		mongodb_statefulset.Spec = appv1.StatefulSetSpec{
			ServiceName: mongodb_statefulset.Name,
			Replicas:    pointer.Int32(growi.Spec.Mongo_db_replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": mongodb_statefulset.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": mongodb_statefulset.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mongo",
							Image: "mongo:" + growi.Spec.Mongo_db_version,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 27017,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      mongodb_statefulset.Name + "-data",
									MountPath: "/data/db",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: mongodb_statefulset.Name + "-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &growi.Spec.Mongo_db_storageClass,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						VolumeName: mongodb_statefulset.Name + "-data",
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(growi.Spec.Mongo_db_storageRequest),
							},
						},
					},
				},
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Growi_deployment", op)
	}

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateMongodbService(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	mongodb_service := &corev1.Service{}
	mongodb_service.SetName("growi-" + growi.Name + "-db-service")
	mongodb_service.SetNamespace(growi.Spec.Mongo_db_namespace)
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, mongodb_service, func() error {
		mongodb_service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": "growi-" + growi.Name + "-db"},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       27017,
					TargetPort: intstr.FromInt(27017),
				},
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Growi_deployment", op)
	}
	return ctrl.Result{}, nil
}
