package api

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

type ExperimentCreateRequest struct {
	Name             string          `json:"name"`
	ArtifactLocation string          `json:"artifact_location"`
	Tags             []ExperimentTag `json:"tags"`
}

type ExperimentCreateResponse struct {
	ID string `json:"experiment_id"`
}

type ExperimentUpdateRequest struct {
	ID   string `json:"experiment_id"`
	Name string `json:"new_name"`
}

type ExperimentGetResponse struct {
	Experiment Experiment `json:"experiment"`
}

type ExperimentDeleteRequest ExperimentCreateResponse

type ExperimentRestoreRequest ExperimentCreateResponse

type ExperimentSetTagRequest struct {
	ID string `json:"experiment_id"`
	ExperimentTag
}

type ExperimentSearchRequest struct {
	MaxResults int64    `json:"max_results"`
	PageToken  string   `json:"page_token"`
	Filter     string   `json:"filter"`
	OrderBy    []string `json:"order_by"`
	ViewType   ViewType `json:"view_type"`
}

type ExperimentSearchResponse struct {
	Experiments   []Experiment `json:"experiments"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}
type RunCreateRequest struct {
	ExperimentID string   `json:"experiment_id"`
	UserID       string   `json:"user_id"`
	Name         string   `json:"run_name"`
	StartTime    int64    `json:"start_time"`
	Tags         []RunTag `json:"tags"`
}

type RunCreateResponse struct {
	Run Run `json:"run"`
}

type RunUpdateRequest struct {
	ID      string    `json:"run_id"`
	UUID    string    `json:"run_uuid"`
	Name    string    `json:"run_name"`
	Status  RunStatus `json:"status"`
	EndTime int64     `json:"end_time"`
}

type RunUpdateResponse struct {
	RunInfo RunInfo `json:"run_info"`
}

type RunGetRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
}

type RunGetResponse RunCreateResponse

type RunSearchRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	Filter        string   `json:"filter"`
	ViewType      ViewType `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
	OrderBy       []string `json:"order_by"`
	PageToken     string   `json:"page_token"`
}

type RunSearchResponse struct {
	Runs          []Run  `json:"runs"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type RunRestoreRequest RunGetRequest

type RunDeleteRequest RunGetRequest

type RunLogParamRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	RunParam
}

type RunLogMetricRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	Metric
}

type RunLogBatchRequest struct {
	ID string `json:"run_id"`
	RunData
}

type RunSetTagRequest struct {
	ID   string `json:"run_id"`
	UUID string `json:"run_uuid"`
	RunTag
}

type RunDeleteTagRequest struct {
	ID  string `json:"run_id"`
	Key string `json:"key"`
}

type ArtifactListResponse struct {
	RootURI       string `json:"root_uri"`
	Files         []File `json:"files"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type MetricsGetHistoryResponse struct {
	Metrics []Metric `json:"metrics"`
}

type MetricsGetHistoriesRequest struct {
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
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
	Step      int64       `json:"step"`
}

type File struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}
