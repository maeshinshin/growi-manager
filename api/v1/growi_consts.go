package v1

const (
	WaitingOtherProcessGrowiApp GrowiAppStatusType = "WaitingOtherProcess"
	StartingGrowiApp            GrowiAppStatusType = "Starting"
	RunningGrowiApp             GrowiAppStatusType = "Running"
	FailedtoStartGrowiApp       GrowiAppStatusType = "FailedToStart"

	WaitingOtherProcessMongodb MongodbStatusType = "WaitingOtherProcess"
	StartingMongodb            MongodbStatusType = "Starting"
	RunningMongodb             MongodbStatusType = "Running"
	FailedtoCreateMongodb      MongodbStatusType = "FailedToStart"

	WaitingOtherProcessElasticsearch ElasticsearchStatusType = "WaitingOtherProcess"
	StartingElasticsearch            ElasticsearchStatusType = "Starting"
	RunningElasticsearch             ElasticsearchStatusType = "Running"
	FailedtoCreateElasticsearch      ElasticsearchStatusType = "FailedToStart"
)
