package models

// Status represents Status type which is using in different models.
type Status string

// Supported list of statuses.
const (
	StatusRunning   Status = "RUNNING"
	StatusScheduled Status = "SCHEDULED"
	StatusFinished  Status = "FINISHED"
	StatusFailed    Status = "FAILED"
	StatusKilled    Status = "KILLED"
)

// LifecycleStage represents entity stage
type LifecycleStage string

// Supported list of stages.
const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)
