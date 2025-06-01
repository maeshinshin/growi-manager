package controller

import (
	"fmt"

	growiappv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func getGrowiappImage(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s:%s", GROWI_APP_IMAGE, growi.Spec.GrowiAppSpec.Version)
}

func getGrowiappSecretName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-growiapp-secret", growi.Name)
}

func getGrowiappServiceName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-growiapp-service", growi.Name)
}

func getGrowiappDeploymentName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-growiapp-deployment", growi.Name)
}

func getMongodbImage(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s:%s", MONGODB_IMAGE, growi.Spec.MongodbSpec.Version)
}

func getMongodbSecretName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-secret", growi.Name)
}

func getMongodbJobName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-job", growi.Name)
}

func getMongodbStatefulSetName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-statefulset", growi.Name)
}

func getMongodbPersistentVolumeClaimName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-pvc", growi.Name)
}

func getMongodbPersistentVolumeClaimNameOfZero(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-%s-0", getMongodbPersistentVolumeClaimName(growi), getMongodbStatefulSetName(growi))
}

func getMongodbHeadlessServiceName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-headless-service", growi.Name)
}

func getMongodbServiceName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-service", growi.Name)
}

func getMongodbPodOfZeroFQDN(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-0.%s.%s.svc.cluster.local", getMongodbStatefulSetName(growi), getMongodbHeadlessServiceName(growi), growi.Namespace)
}

func getMongodbServiceFQDN(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", getMongodbServiceName(growi), growi.Namespace)
}

func getMongodbURI(growi growiappv1.Growi, username, pass string) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:27017/growi?authSource=admin", username, pass, getMongodbServiceFQDN(growi))
}

func getElasticsearchImage(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s:%s", ELASTICSEARCH_IMAGE, growi.Spec.ElasticsearchSpec.Version)
}

func getElasticsearchHeadlessServiceName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-headless-service", growi.Name)
}

func getElasticsearchHeadlessServiceFQDN(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-headless-service.%s.svc.cluster.local", growi.Name, growi.Namespace)
}

func getElasticsearchServiceName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-service", growi.Name)
}

func getElasticsearchServiceFQDN(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-service.%s.svc.cluster.local", growi.Name, growi.Namespace)
}

func getElasticsearchURI(growi growiappv1.Growi) string {
	return "http://" + getElasticsearchServiceFQDN(growi) + ":9200/growi"
}

func getElasticsearchStatefulSetName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-statefulset", growi.Name)
}

func getElasticsearchDataPersistentVolumeClaimName(growi growiappv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-data-pvc", growi.Name)
}

func getElasticsearchHostList(growi growiappv1.Growi) string {
	var hostList string
	elasticsearchStatefulSetName := getElasticsearchStatefulSetName(growi)
	// getElasticsearchHeadlessServiceName := getElasticsearchHeadlessServiceName(*growi)
	var i int
	for i = range int(growi.Spec.ElasticsearchSpec.Replicas) - 1 {
		// hostList += fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local,", elasticsearchStatefulSetName, i, getElasticsearchHeadlessServiceName, growi.Namespace)
		hostList += fmt.Sprintf("%s-%d,", elasticsearchStatefulSetName, i)
	}
	// hostList += fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", elasticsearchStatefulSetName, i+1, getElasticsearchHeadlessServiceName, growi.Namespace)
	hostList += fmt.Sprintf("%s-%d", elasticsearchStatefulSetName, i+1)
	return hostList
}

func getLabels(growi growiappv1.Growi, component component) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "growi",
		"app.kubernetes.io/instance":   growi.Name,
		"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
		"app.kubernetes.io/component":  string(component),
	}
}
