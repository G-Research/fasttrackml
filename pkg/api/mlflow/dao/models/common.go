package models

type Status string

const (
	StatusRunning   Status = "RUNNING"
	StatusScheduled Status = "SCHEDULED"
	StatusFinished  Status = "FINISHED"
	StatusFailed    Status = "FAILED"
	StatusKilled    Status = "KILLED"
)

type LifecycleStage string

const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)
