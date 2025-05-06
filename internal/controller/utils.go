package controller

import (
	"fmt"

	gmv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func getGrowiDeploymentName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-deployment", growi.Name)
}

func getGrowiServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-service", growi.Name)
}

func getMongoDBSecretName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-secret", growi.Name)
}

func getMongoServiceName(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-service", growi.Name)
}

func getMongoServiceFQDN(growi gmv1.Growi) string {
	return fmt.Sprintf("%s-mongodb-service.%s.svc.cluster.local", growi.Name, growi.Namespace)
}

func getMongoURI(growi gmv1.Growi, username, pass string) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:27017/growi?authSource=admin", username, pass, getMongoServiceFQDN(growi))
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
