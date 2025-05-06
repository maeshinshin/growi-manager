package v1

const (
	WaitingOtherProcessMongoDBSecret MongoDBSecretStatusType = "WaitingOtherProcess"
	CreatingMongoDBSecret            MongoDBSecretStatusType = "Creating"
	ExistMongoDBSecret               MongoDBSecretStatusType = "Exist"
	FailedtoCreateMongoDBSecret      MongoDBSecretStatusType = "FailedtoCreate"

	WaitingOtherProcessGrowiApp GrowiAppStatusType = "WaitingOtherProcess"
	StartingGrowiApp            GrowiAppStatusType = "Starting"
	RunningGrowiApp             GrowiAppStatusType = "Running"
	FailedtoStartGrowiApp       GrowiAppStatusType = "FailedToStart"

	WaitingOtherProcessMongoDB MongoDBStatusType = "WaitingOtherProcess"
	StartingMongoDB            MongoDBStatusType = "Starting"
	RunningMongoDB             MongoDBStatusType = "Running"
	FailedtoStartMongoDB       MongoDBStatusType = "FailedToStart"

	WaitingOtherProcessElasticSearch ElasticSearchStatusType = "WaitingOtherProcess"
	StartingElasticSearch            ElasticSearchStatusType = "Starting"
	RunningElasticSearch             ElasticSearchStatusType = "Running"
	FailedtoStartElasticSearch       ElasticSearchStatusType = "FailedToStart"
)
