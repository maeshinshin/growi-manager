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
	"fmt"
	"strings"

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
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

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

	logger.Info("Reconciliation triggered")

	var growi maeshinshingithubiov1.Growi
	if err := r.Get(ctx, req.NamespacedName, &growi); err != nil {
		logger.Error(err, "Fetch Growi object: failed")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Fetch Growi object: succeeded", "growi", growi.Name, "createdAt", growi.CreationTimestamp)

	if res, err := r.CreateOrUpdateGrowiNamespace(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateMongodbNamespace(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateElasticsearchNamespace(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateMongodbService(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateMongodbStatefulSet(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateElasticsearchService(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateElasticsearchStatefulSet(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateGrowiService(ctx, req, growi, logger); err != nil {
		return res, err
	}

	if res, err := r.CreateOrUpdateGrowiStatefulSet(ctx, req, growi, logger); err != nil {
		return res, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrowiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&maeshinshingithubiov1.Growi{}).
		Owns(&appv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

// make Namespace for growi-app
func (r *GrowiReconciler) CreateOrUpdateGrowiNamespace(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	growi_namespace := &corev1.Namespace{}
	growi_namespace.SetName(growi.Spec.Growi_app_namespace)
	logger.Info("CreateOrUpdateGrowiNamespace: start", "Namespace", growi_namespace.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, growi_namespace, func() error {
		if growi_namespace.Labels == nil {
			growi_namespace.SetLabels(map[string]string{"app": growi.Name})
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateGrowiNamespace: Operation Finished", "Operation", op)
	}

	logger.Info("CreateOrUpdateGrowiNamespace: succeeded", "Namespace", growi_namespace.Name)

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateMongodbNamespace(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	mongo_db_namespace := &corev1.Namespace{}
	mongo_db_namespace.SetName(growi.Spec.Mongo_db_namespace)
	logger.Info("CreateOrUpdateMongodbNamespace: start", "Namespace", mongo_db_namespace.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, mongo_db_namespace, func() error {
		if mongo_db_namespace.Labels == nil {
			mongo_db_namespace.SetLabels(map[string]string{"app": growi.Name})
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateMongodbNamespace: Operation Finished", "", op)
	}

	logger.Info("CreateOrUpdateMongodbNamespace: succeeded", "Namespace", mongo_db_namespace.Name)
	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateElasticsearchNamespace(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	elasticsearch_namespace := &corev1.Namespace{}
	elasticsearch_namespace.SetName(growi.Spec.Elasticsearch_namespace)
	logger.Info("CreateOrUpdateElasticsearchNamespace: start", "Namespace", elasticsearch_namespace.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, elasticsearch_namespace, func() error {
		if elasticsearch_namespace.Labels == nil {
			elasticsearch_namespace.SetLabels(map[string]string{"app": growi.Name})
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateElasticsearchNamespace: Operation Finished", "Operation", op)
	}

	logger.Info("CreateOrUpdateElasticsearchNamespace: succeeded", "Namespace", elasticsearch_namespace.Name)

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateGrowiStatefulSet(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	// Create Deployment object if not exists
	growi_statefulset := &appv1.StatefulSet{}
	growi_statefulset.SetNamespace(growi.Spec.Growi_app_namespace)
	growi_statefulset.SetName("growi-app-" + growi.Name)
	logger.Info("CreateOrUpdateGrowiStatefulSet: start", "StatefulSet", growi_statefulset.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, growi_statefulset, func() error {
		growi_statefulset.Spec = appv1.StatefulSetSpec{
			ServiceName: "growi-app-service" + growi.Name,
			Replicas:    pointer.Int32(growi.Spec.Growi_replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": growi_statefulset.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": growi_statefulset.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "growi",
							Image: "weseek/growi:" + growi.Spec.Growi_version,
							Command: []string{
								"/bin/sh",
								"-c",
								"yarn migrate && node -r dotenv-flow/config --expose_gc dist/server/app.js",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "growi-app-pvc-" + growi.Name,
									MountPath: "/data",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MONGO_URI",
									Value: "mongodb://growi-db-service-" + growi.Name + "." + growi.Spec.Mongo_db_namespace + ".svc.cluster.local" + ":27017/growi",
								},
								{
									Name:  "ELASTICSEARCH_URI",
									Value: "http://growi-es-service-" + growi.Name + "." + growi.Spec.Elasticsearch_namespace + ".svc.cluster.local" + ":9200/growi",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "growi-app-pvc-" + growi.Name,
						Namespace: growi.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &growi.Spec.Elasticsearch_storageClass,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(growi.Spec.Growi_storageRequest),
							},
						},
					},
				},
			},
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateGrowiStatefulSet: Operation Finished", "Operation", op)
	}

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateGrowiService(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	growi_service := &corev1.Service{}
	growi_service.SetName("growi-app-service-" + growi.Name)
	growi_service.SetNamespace(growi.Spec.Growi_app_namespace)
	logger.Info("CreateOrUpdateGrowiService: start", "Service", growi_service.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, growi_service, func() error {
		growi_service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": "growi-app-" + growi.Name},
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
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateGrowiService: Operation Finished", "Operation", op)
	}
	logger.Info("CreateOrUpdateGrowiService: succeeded", "service", growi_service)
	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateMongodbStatefulSet(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	// Create Deployment object if not exists
	mongodb_statefulset := &appv1.StatefulSet{}
	mongodb_statefulset.SetNamespace(growi.Spec.Mongo_db_namespace)
	mongodb_statefulset.SetName("growi-db-" + growi.Name)
	logger.Info("CreateOrUpdateMongodbStatefulSet: start", "StatefulSet", mongodb_statefulset.Name)
	members := make([]string, growi.Spec.Mongo_db_replicas)
	for i := 0; i < int(growi.Spec.Mongo_db_replicas); i++ {
		host := fmt.Sprintf(mongodb_statefulset.Name+"-%d."+"growi-db-service-"+growi.Name+"."+growi.Spec.Mongo_db_namespace+".svc.cluster.local:27017", i)
		member := fmt.Sprintf("{_id: %d, host: \"%s\"}", i, host)
		members[i] = member
	}
	membersString := strings.Join(members, ", ")
	mongoCommand := fmt.Sprintf("mongosh --eval 'rs.initiate({_id: \"%s\", members: [%s]})';", growi.Name, membersString)

	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, mongodb_statefulset, func() error {
		mongodb_statefulset.Spec = appv1.StatefulSetSpec{
			ServiceName: "growi-db-service-" + growi.Name,
			Replicas:    pointer.Int32(growi.Spec.Mongo_db_replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": mongodb_statefulset.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": mongodb_statefulset.Name},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "chmod-er",
							Image: "busybox:latest",
							Command: []string{
								"/bin/chown",
								"-R",
								"1000",
								"/data/db",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "growi-db-pvc-" + growi.Name,
									MountPath: "/data/db",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mongo",
							Image: "mongo:" + growi.Spec.Mongo_db_version,
							Command: []string{
								"/bin/sh",
								"-c",
							},
							Args: []string{
								fmt.Sprintf(
									"mongod --replSet=%s --bind_ip_all & sleep 10 && if [ $(hostname) = '%s-%d' ]; then %s fi && wait",
									growi.Name,
									mongodb_statefulset.Name,
									growi.Spec.Mongo_db_replicas-1,
									mongoCommand,
								),
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 27017,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "growi-db-pvc-" + growi.Name,
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
						Name:      "growi-db-pvc-" + growi.Name,
						Namespace: growi.Spec.Mongo_db_namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &growi.Spec.Mongo_db_storageClass,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
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
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateMongodbStatefulSet: Operation Finished", "Operation", op)
	}

	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateMongodbService(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	mongodb_service := &corev1.Service{}
	mongodb_service.SetName("growi-db-service-" + growi.Name)
	mongodb_service.SetNamespace(growi.Spec.Mongo_db_namespace)
	logger.Info("CreateOrUpdateMongodbService: start", "Service", mongodb_service.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, mongodb_service, func() error {
		mongodb_service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": "growi-db-" + growi.Name},
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
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateMongodbService: Operation Finished", "Operation", op)
	}
	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateElasticsearchStatefulSet(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	elasticsearch_statefulset := &appv1.StatefulSet{}
	elasticsearch_statefulset.SetName("growi-es-" + growi.Name)
	elasticsearch_statefulset.SetNamespace(growi.Spec.Elasticsearch_namespace)
	logger.Info("CreateOrUpdateElasticsearchStatefulSet: start", "StatefulSet", elasticsearch_statefulset.Name)
	if op, err := ctrl.CreateOrUpdate(ctx, r.Client, elasticsearch_statefulset, func() error {
		elasticsearch_statefulset.Spec = appv1.StatefulSetSpec{
			ServiceName: "growi-es-service-" + growi.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": elasticsearch_statefulset.Name},
			},
			Replicas: pointer.Int32(growi.Spec.Elasticsearch_replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": elasticsearch_statefulset.Name},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "install-plugins",
							Image: "docker.elastic.co/elasticsearch/elasticsearch:" + growi.Spec.Elasticsearch_version,
							Command: []string{
								"/bin/sh",
								"-c",
								`bin/elasticsearch-plugin remove --purge analysis-kuromoji;bin/elasticsearch-plugin install --batch analysis-kuromoji;bin/elasticsearch-plugin remove --purge analysis-icu;bin/elasticsearch-plugin install --batch analysis-icu`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "growi-es-plugins-pvc-" + growi.Name,
									MountPath: "/usr/share/elasticsearch/plugins",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "elasticsearch",
							Image: "docker.elastic.co/elasticsearch/elasticsearch:" + growi.Spec.Elasticsearch_version,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9200,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "growi-es-pvc-" + growi.Name,
									MountPath: "/usr/share/elasticsearch/data",
								},
								{
									Name:      "growi-es-plugins-pvc-" + growi.Name,
									MountPath: "/usr/share/elasticsearch/plugins",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "network.host",
									Value: "0.0.0.0",
								},
								{
									Name:  "http.cors.enabled",
									Value: "true",
								},
								{
									Name:  "http.cors.allow-origin",
									Value: "\"*\"",
								},
								{
									Name:  "xpack.ml.enabled",
									Value: "false",
								},
								{
									Name:  "bootstrap.memory_lock",
									Value: "true",
								},
								{
									Name:  "ES_JAVA_OPTS",
									Value: "-Xms256m -Xmx256m",
								},
								{
									Name:  "LOG4J_FORMAT_MSG_NO_LOOKUPS",
									Value: "true",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "growi-es-pvc-" + growi.Name,
						Namespace: growi.Spec.Elasticsearch_namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &growi.Spec.Elasticsearch_storageClass,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(growi.Spec.Elasticsearch_storageRequest),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "growi-es-plugins-pvc-" + growi.Name,
						Namespace: growi.Spec.Elasticsearch_namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &growi.Spec.Elasticsearch_storageClass,
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	} else if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateElasticsearchStatefulSet: Operation Finished", "Operation", op)
	}

	logger.Info("CreateOrUpdateElasticsearchStatefulSet: succeeded", "StatefulSet", elasticsearch_statefulset.Name)
	return ctrl.Result{}, nil
}

func (r *GrowiReconciler) CreateOrUpdateElasticsearchService(ctx context.Context, req ctrl.Request, growi maeshinshingithubiov1.Growi, logger logr.Logger) (ctrl.Result, error) {
	elasticsearch_service := &corev1.Service{}
	elasticsearch_service.SetName("growi-es-service-" + growi.Name + "")
	elasticsearch_service.SetNamespace(growi.Spec.Elasticsearch_namespace)
	logger.Info("CreateOrUpdateElasticsearchService: start", "Service", elasticsearch_service.Name)
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, elasticsearch_service, func() error {
		elasticsearch_service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": "growi-es-" + growi.Name},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       9200,
					TargetPort: intstr.FromInt(9200),
				},
			},
		}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("CreateOrUpdateElasticsearchService: Operation Finished", "Operation", op)
	}
	return ctrl.Result{}, nil
}
