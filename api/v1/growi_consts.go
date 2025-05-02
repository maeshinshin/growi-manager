package v1

const (
	SettingGrowiApp       GrowiStatusType = "Setting"
	ReadyGrowiApp         GrowiStatusType = "Ready"
	FailedtoStartGrowiApp GrowiStatusType = "FailedToStart"

	SettingMongoDB       MongoDBStatusType = "Setting"
	ReadyMongoDB         MongoDBStatusType = "Ready"
	FailedtoStartMongoDB MongoDBStatusType = "FailedToStart"

	SettingElasticSearch       ElasticSearchStatusType = "Setting"
	ReadyElasticSearch         ElasticSearchStatusType = "Ready"
	FailedtoStartElasticSearch ElasticSearchStatusType = "FailedToStart"
)
