package models

// LifecycleStage represents entity stage
type LifecycleStage string

// Supported list of stages.
const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)
