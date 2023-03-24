package mlflow

import "fmt"

// TODO conversion methods to/from model
// or maybe even better, compatibility between both?

type ErrorResponse struct {
	ErrorCode ErrorCode `json:"error_code"`
	Message   string    `json:"message"`
}

type ErrorCode string

const (
	ErrorCodeInternalError          = "INTERNAL_ERROR"
	ErrorCodeTemporarilyUnavailable = "TEMPORARILY_UNAVAILABLE"
	ErrorCodeBadRequest             = "BAD_REQUEST"
	ErrorCodeInvalidParameterValue  = "INVALID_PARAMETER_VALUE"
	ErrorCodeEndpointNotFound       = "ENDPOINT_NOT_FOUND"
	ErrorCodeInvalidState           = "INVALID_STATE"
	ErrorCodeResourceAlreadyExists  = "RESOURCE_ALREADY_EXISTS"
	ErrorCodeResourceDoesNotExist   = "RESOURCE_DOES_NOT_EXIST"
)

func NewError(e ErrorCode, msg string, args ...any) *ErrorResponse {
	return &ErrorResponse{
		ErrorCode: e,
		Message:   fmt.Sprintf(msg, args...),
	}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}

type CreateExperimentRequest struct {
	Name             string          `json:"name"`
	ArtifactLocation string          `json:"artifact_location"`
	Tags             []ExperimentTag `json:"tags"`
}

type CreateExperimentResponse struct {
	ID string `json:"experiment_id"`
}

type UpdateExperimentRequest struct {
	ID   string `json:"experiment_id"`
	Name string `json:"new_name"`
}

type GetExperimentResponse struct {
	Experiment Experiment `json:"experiment"`
}

type DeleteExperimentRequest CreateExperimentResponse

type RestoreExperimentRequest CreateExperimentResponse

type SetExperimentTagRequest struct {
	ID string `json:"experiment_id"`
	ExperimentTag
}

type SearchExperimentsRequest struct {
	MaxResults int64    `json:"max_results" query:"max_results"`
	PageToken  string   `json:"page_token"  query:"page_token"`
	Filter     string   `json:"filter"      query:"filter"`
	OrderBy    []string `json:"order_by"    query:"order_by"`
	ViewType   ViewType `json:"view_type"   query:"view_type"`
}

type SearchExperimentsResponse struct {
	Experiments   []Experiment `json:"experiments"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}
type CreateRunRequest struct {
	ExperimentID string   `json:"experiment_id"`
	UserID       string   `json:"user_id"`
	Name         string   `json:"run_name"`
	StartTime    int64    `json:"start_time"`
	Tags         []RunTag `json:"tags"`
}

type CreateRunResponse struct {
	Run Run `json:"run"`
}

type UpdateRunRequest struct {
	ID      string    `json:"run_id"`
	UUID    string    `json:"run_uuid"`
	Name    string    `json:"run_name"`
	Status  RunStatus `json:"status"`
	EndTime int64     `json:"end_time"`
}

type UpdateRunResponse struct {
	RunInfo RunInfo `json:"run_info"`
}

type RunGetRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

type GetRunResponse CreateRunResponse

type SearchRunsRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	Filter        string   `json:"filter"`
	ViewType      ViewType `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
	OrderBy       []string `json:"order_by"`
	PageToken     string   `json:"page_token"`
}

type SearchRunsResponse struct {
	Runs          []Run  `json:"runs"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type RestoreRunRequest RunGetRequest

type DeleteRunRequest RunGetRequest

type LogParamRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	RunParam
}

type LogMetricRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	Metric
}

type LogBatchRequest struct {
	ID string `json:"run_id"`
	RunData
}

type SetRunTagRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	RunTag
}

type DeleteRunTagRequest struct {
	ID  string `json:"run_id"`
	Key string `json:"key"`
}

type ListArtifactsResponse struct {
	RootURI       string `json:"root_uri"`
	Files         []File `json:"files"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type GetMetricHistoryResponse struct {
	Metrics []Metric `json:"metrics"`
}

type GetMetricHistoriesRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	RunIDs        []string `json:"run_ids"`
	MetricKeys    []string `json:"metric_keys"`
	ViewType      ViewType `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ViewType string

const (
	ViewTypeActiveOnly  ViewType = "ACTIVE_ONLY"
	ViewTypeDeletedOnly ViewType = "DELETED_ONLY"
	ViewTypeAll         ViewType = "ALL"
)

type PageToken struct {
	Offset int32 `json:"offset"`
}

type Experiment struct {
	ID               string          `json:"experiment_id"`
	Name             string          `json:"name"`
	ArtifactLocation string          `json:"artifact_location"`
	LifecycleStage   LifecycleStage  `json:"lifecycle_stage"`
	LastUpdateTime   int64           `json:"last_update_time"`
	CreationTime     int64           `json:"creation_time"`
	Tags             []ExperimentTag `json:"tags"`
}

type ExperimentTag KV

type Run struct {
	Info RunInfo `json:"info"`
	Data RunData `json:"data"`
}

type RunInfo struct {
	UUID           string         `json:"run_uuid"`
	Name           string         `json:"run_name"`
	ExperimentID   string         `json:"experiment_id"`
	UserID         string         `json:"user_id,omitempty"`
	Status         RunStatus      `json:"status"`
	StartTime      int64          `json:"start_time"`
	EndTime        int64          `json:"end_time,omitempty"`
	ArtifactURI    string         `json:"artifact_uri,omitempty"`
	LifecycleStage LifecycleStage `json:"lifecycle_stage"`
	ID             string         `json:"run_id"`
}

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

type RunData struct {
	Metrics []Metric   `json:"metrics,omitempty"`
	Params  []RunParam `json:"params,omitempty"`
	Tags    []RunTag   `json:"tags,omitempty"`
}

type RunParam KV

type RunTag KV

type Metric struct {
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

type File struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}
