package controller

import "time"

const (
	GROWI_APP_IMAGE     = "weseek/growi"
	MONGODB_IMAGE       = "mongo"
	ELASTICSEARCH_IMAGE = "docker.elastic.co/elasticsearch/elasticsearch"

	FIELDMANAGER_NAME = "growi-controller"
	FINALIZER_NAME    = "growi.apps.maesh.dev/finalizer"

	// Component labels
	COMPONENT_INIT_MONGODB  component = "init-mongodb"
	COMPONENT_GROWIAPP      component = "growi"
	COMPONENT_MONGODB       component = "mongodb"
	COMPONENT_ELASTICSEARCH component = "elasticsearch"

	REQUEUE_INTERVAL = time.Second * 10

	CREDENTIAL_CHARACTER_SET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)
