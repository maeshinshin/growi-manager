package controller

import (
	"fmt"

	gmv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func getGrowiAppImage(growi gmv1.Growi) string {
	return fmt.Sprintf("%s:%s", GROWI_APP_IMAGE, growi.Spec.GrowiAppSpec.Version)
}

func getGrowiDeploymentName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-deployment", growi.Name)
}

func getGrowiServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-service", growi.Name)
}

func getMongodbImage(growi gmv1.Growi) string {
	return fmt.Sprintf("%s:%s", MONGODB_IMAGE, growi.Spec.MongodbSpec.Version)
}

func getMongodbSecretName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-secret", growi.Name)
}

func getMongodbJobName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-job", growi.Name)
}

func getMongodbStatefulSetName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-statefulset", growi.Name)
}

func getMongodbPersistentVolumeClaimName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-pvc", growi.Name)
}

func getMongodbHeadlessServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-headless-service", growi.Name)
}

func getMongodbServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-service", growi.Name)
}

func getMongodbPodOfZeroFQDN(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-0.%s.%s.svc.cluster.local", getMongodbStatefulSetName(growi), getMongodbHeadlessServiceName(growi), growi.Namespace)
}

func getMongodbServiceFQDN(growi gmv1.Growi) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", getMongodbServiceName(growi), growi.Namespace)
}

func getMongodbURI(growi gmv1.Growi, username, pass string) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:27017/growi?authSource=admin", username, pass, getMongodbServiceFQDN(growi))
}

func getElasticSearchImage(growi gmv1.Growi) string {
	return fmt.Sprintf("%s:%s", ELASTICSEARCH_IMAGE, growi.Spec.ElasticsearchSpec.Version)
}

func getElasticSearchServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-service", growi.Name)
}

func getElasticSearchServiceFQDN(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-elasticsearch-service.%s.svc.cluster.local", growi.Name, growi.Namespace)
}

func getElasticSearchURI(growi gmv1.Growi) string {
	return "http://" + getElasticSearchServiceFQDN(growi) + ":9200/growi"
}

func getLabels(growi gmv1.Growi, component component) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "growi",
		"app.kubernetes.io/instance":   growi.Name,
		"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
		"app.kubernetes.io/component":  string(component),
	}
}
