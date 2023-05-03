package request

type RunStatus string

const (
	RunStatusRunning   RunStatus = "RUNNING"
	RunStatusScheduled RunStatus = "SCHEDULED"
	RunStatusFinished  RunStatus = "FINISHED"
	RunStatusFailed    RunStatus = "FAILED"
	RunStatusKilled    RunStatus = "KILLED"
)

type LifecycleStage string

const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)

type ViewType string

const (
	ViewTypeActiveOnly  ViewType = "ACTIVE_ONLY"
	ViewTypeDeletedOnly ViewType = "DELETED_ONLY"
	ViewTypeAll         ViewType = "ALL"
)

type PageToken struct {
	Offset int32 `json:"offset"`
}
