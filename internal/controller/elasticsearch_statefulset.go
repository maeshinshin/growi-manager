package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileElasticsearchStatefulSet(ctx context.Context, growi *growiv1.Growi) error {
	var err error
	logger := logf.FromContext(ctx)
	elasticsearchStatefulSetName := getElasticsearchStatefulSetName(*growi)
	elasticsearchHeadlessServiceName := getElasticsearchHeadlessServiceName(*growi)
	elasticsearchStatefulsetLabels := getLabels(*growi, COMPONENT_ELASTICSEARCH)
	elasticsearchNodeList := getElasticsearchHostList(growi)
	elasticsearchDataPersistentVolumeClaimName := getElasticsearchDataPersistentVolumeClaimName(*growi)

	// Chech if the Elasticsearch statefulset already exists
	currElasticsearchStatefulSet := &appsv1.StatefulSet{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      elasticsearchStatefulSetName,
		Namespace: growi.Namespace,
	}, currElasticsearchStatefulSet)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get Elasticsearch statefulset")
		return err
	}

	// Create the Elasticsearch statefulset
	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(ctx, growi); err != nil {
		return err
	}

	elasticsearchStatefulSet := appsv1apply.StatefulSet(
		elasticsearchStatefulSetName,
		growi.Namespace,
	).
		WithLabels(elasticsearchStatefulsetLabels).
		WithOwnerReferences(ownerRef).
		WithSpec(
			appsv1apply.StatefulSetSpec().
				WithServiceName(elasticsearchHeadlessServiceName).
				WithReplicas(growi.Spec.ElasticsearchSpec.Replicas).
				WithPodManagementPolicy(appsv1.ParallelPodManagement).
				WithSelector(
					metav1apply.LabelSelector().
						WithMatchLabels(elasticsearchStatefulsetLabels),
				).
				WithTemplate(
					corev1apply.PodTemplateSpec().
						WithLabels(elasticsearchStatefulsetLabels).
						WithSpec(
							corev1apply.PodSpec().
								WithInitContainers(
									corev1apply.Container().
										WithName("sysctl").
										WithImage("busybox").
										WithCommand("sysctl", "-w", "vm.max_map_count=262144").
										WithSecurityContext(
											corev1apply.SecurityContext().
												WithPrivileged(true),
										),
									corev1apply.Container().
										WithName("install-plugins").
										WithImage(getElasticsearchImage(*growi)).
										WithCommand(
											"/bin/sh",
											"-c",
											"rm /usr/share/elasticsearch/plugins/* -rf && bin/elasticsearch-plugin install analysis-icu && bin/elasticsearch-plugin install analysis-kuromoji",
										).
										WithVolumeMounts(
											corev1apply.VolumeMount().
												WithName("elasticsearch-plugin").
												WithMountPath("/usr/share/elasticsearch/plugins"),
										),
								).
								WithContainers(
									corev1apply.Container().
										WithName("elasticsearch").
										WithImage(getElasticsearchImage(*growi)).
										WithPorts(
											corev1apply.ContainerPort().
												WithName("http").
												WithContainerPort(9200).
												WithProtocol(corev1.ProtocolTCP),
											corev1apply.ContainerPort().
												WithName("transport").
												WithContainerPort(9300).
												WithProtocol(corev1.ProtocolTCP),
										).
										WithEnv(
											corev1apply.EnvVar().
												WithName("ES_JAVA_OPTS").
												WithValue("-Xms512m -Xmx512m"),
											corev1apply.EnvVar().
												WithName("cluster.name").
												WithValue(elasticsearchStatefulSetName),
											corev1apply.EnvVar().
												WithName("network.host").
												WithValue("0.0.0.0"),
											corev1apply.EnvVar().
												WithName("http.cors.enabled").
												WithValue("true"),
											corev1apply.EnvVar().
												WithName("http.cors.allow-origin").
												WithValue("\"*\""),
											corev1apply.EnvVar().
												WithName("discovery.seed_hosts").
												WithValue(elasticsearchNodeList),
											corev1apply.EnvVar().
												WithName("cluster.initial_master_nodes").
												WithValue(elasticsearchNodeList),
											corev1apply.EnvVar().
												WithName("node.roles").
												WithValue("[master, data, ingest]"),
											corev1apply.EnvVar().
												WithName("bootstrap.memory_lock").
												WithValue("false"),
											corev1apply.EnvVar().
												WithName("xpack.security.enabled").
												WithValue("false"),
										).
										WithVolumeMounts(
											corev1apply.VolumeMount().
												WithName("elasticsearch-plugin").
												WithMountPath("/usr/share/elasticsearch/plugins"),
											corev1apply.VolumeMount().
												WithName(elasticsearchDataPersistentVolumeClaimName).
												WithMountPath("/usr/share/elasticsearch/data"),
										).
										WithLivenessProbe(
											corev1apply.Probe().
												WithHTTPGet(
													corev1apply.HTTPGetAction().
														WithPath("/").
														WithPort(intstr.FromString("http")).
														WithScheme(corev1.URISchemeHTTP),
												).
												WithInitialDelaySeconds(60).
												WithTimeoutSeconds(5).
												WithPeriodSeconds(10).
												WithSuccessThreshold(1).
												WithFailureThreshold(5),
										).
										WithReadinessProbe(
											corev1apply.Probe().
												WithHTTPGet(
													corev1apply.HTTPGetAction().
														WithPath("/_cluster/health?local=true&wait_for_status=yellow&timeout=1s").
														WithPort(intstr.FromString("http")).
														WithScheme(corev1.URISchemeHTTP),
												).
												WithInitialDelaySeconds(90).
												WithTimeoutSeconds(5).
												WithPeriodSeconds(10).
												WithSuccessThreshold(1).
												WithFailureThreshold(5),
										).
										WithSecurityContext(
											corev1apply.SecurityContext().
												WithCapabilities(
													corev1apply.Capabilities().
														WithAdd(
															"IPC_LOCK",
														),
												),
										),
								).
								WithVolumes(
									corev1apply.Volume().
										WithName("elasticsearch-plugin").
										WithEmptyDir(corev1apply.EmptyDirVolumeSource()),
								),
						),
				).
				WithVolumeClaimTemplates(
					corev1apply.PersistentVolumeClaim(
						elasticsearchDataPersistentVolumeClaimName,
						growi.Namespace,
					).
						WithLabels(elasticsearchStatefulsetLabels).
						WithSpec(
							corev1apply.PersistentVolumeClaimSpec().
								WithAccessModes(corev1.ReadWriteOnce).
								WithResources(
									corev1apply.VolumeResourceRequirements().
										WithRequests(corev1.ResourceList{
											corev1.ResourceStorage: apiresource.MustParse("10Gi"),
										}),
								).
								WithVolumeMode(corev1.PersistentVolumeFilesystem).
								WithStorageClassName(growi.Spec.StorageClass),
						),
				),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(elasticsearchStatefulSet)
	if err != nil {
		return err
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := appsv1apply.ExtractStatefulSet(currElasticsearchStatefulSet, FIELDMANAGER_NAME)
	if err != nil {
		logger.Error(err, "unable to extract statefulset")
		return err
	}

	if !shouldPatch(ctx, currApplyConfig, elasticsearchStatefulSet) {
		logger.Info("elasticsearch statefulset already exists, skipping creation")
		return nil
	}

	if err := r.updateElasticsearchStatus(ctx, growi, growiv1.StartingElasticsearch); err != nil {
		return err
	}

	logger.Info("Creating or updating elasticsearch statefulset", "name", elasticsearchStatefulSetName)
	if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create elasticsearch statefulset")
		return err
	}

	return nil
}
