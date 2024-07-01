package run

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/database"
)

//nolint:lll
var (
	runOrder      = regexp.MustCompile(`^(attribute|metric|param|tag)s?\.("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)(?i:\s+(ASC|DESC))?$`)
	filterAnd     = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond    = regexp.MustCompile(`^(?:(\w+)\.)?("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'[^']+'(?:,\s*)?)+\)|"[^"]+"|'[^']+'|[\w\.]+)$`)
	filterInGroup = regexp.MustCompile(`,\s*`)
)

// supported expression list.
const (
	InExpression            = "IN"
	NotInExpression         = "NOT IN"
	LikeExpression          = "LIKE"
	ILikeExpression         = "ILIKE"
	EqualExpression         = "="
	NotEqualExpression      = "!="
	LessExpression          = "<"
	LessOrEqualExpression   = "<="
	GraterExpression        = ">"
	GraterOrEqualExpression = ">="
)

// Service provides service layer to work with `run` business logic.
type Service struct {
	logRepository        repositories.LogRepositoryProvider
	tagRepository        repositories.TagRepositoryProvider
	runRepository        repositories.RunRepositoryProvider
	paramRepository      repositories.ParamRepositoryProvider
	metricRepository     repositories.MetricRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
	artifactRepository   repositories.ArtifactRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	tagRepository repositories.TagRepositoryProvider,
	runRepository repositories.RunRepositoryProvider,
	paramRepository repositories.ParamRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
	logRepository repositories.LogRepositoryProvider,
	artifactRepository repositories.ArtifactRepositoryProvider,
) *Service {
	return &Service{
		logRepository:        logRepository,
		tagRepository:        tagRepository,
		runRepository:        runRepository,
		paramRepository:      paramRepository,
		metricRepository:     metricRepository,
		experimentRepository: experimentRepository,
		artifactRepository:   artifactRepository,
	}
}

func (s Service) CreateRun(
	ctx context.Context, ns *models.Namespace, req *request.CreateRunRequest,
) (*models.Run, error) {
	adjustCreateRunRequestForNamespace(ns, req)
	experimentID, err := strconv.ParseInt(req.ExperimentID, 10, 32)
	if err != nil {
		return nil, api.NewBadRequestError("unable to parse experiment id '%s': %s", req.ExperimentID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(experimentID))
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find experiment with id '%s': %s", req.ExperimentID, err)
	}

	run, err := convertors.ConvertCreateRunRequestToDBModel(experiment, req)
	if err != nil {
		return nil, api.NewInternalError("error converting request to actual run model: %s", err)
	}
	if err := s.runRepository.Create(ctx, run); err != nil {
		return nil, api.NewInternalError("error inserting run: %s", err)
	}

	return run, nil
}

func (s Service) UpdateRun(
	ctx context.Context, namespace *models.Namespace, req *request.UpdateRunRequest,
) (*models.Run, error) {
	if err := ValidateUpdateRunRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.GetRunID())
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s'", req.GetRunID())
	}

	run = convertors.ConvertUpdateRunRequestToDBModel(run, req)
	if err := s.runRepository.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := s.runRepository.UpdateWithTransaction(ctx, tx, run); err != nil {
			return err
		}
		if req.Name != "" {
			// TODO:DSuhinin - move "mlflow.runName" to be a constant somewhere.
			// Also, Im not fully sure that this is right place to keep this logic here.
			if err := s.tagRepository.CreateRunTagWithTransaction(
				ctx, tx, run.ID, "mlflow.runName", req.Name,
			); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, api.NewInternalError("unable to update run '%s': %s", run.ID, err)
	}

	return run, nil
}

func (s Service) GetRun(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.GetRunRequest,
) (*models.Run, error) {
	if err := ValidateGetRunRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.GetRunID())
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.GetRunID(), err)
	}
	if run == nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s'", req.GetRunID())
	}

	return run, nil
}

// nolint:gocyclo
// TODO:get back and fix `gocyclo` problem.
func (s Service) SearchRuns(
	ctx context.Context, namespace *models.Namespace, req *request.SearchRunsRequest,
) ([]models.Run, int, int, error) {
	if err := ValidateSearchRunsRequest(req); err != nil {
		return nil, 0, 0, err
	}
	adjustSearchRunsRequestForNamespace(namespace, req)

	// ViewType
	var lifecyleStages []database.LifecycleStage
	switch req.ViewType {
	case request.ViewTypeActiveOnly, "":
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageActive,
		}
	case request.ViewTypeDeletedOnly:
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageDeleted,
		}
	case request.ViewTypeAll:
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageActive,
			database.LifecycleStageDeleted,
		}
	}
	tx := database.DB.Joins(
		"LEFT JOIN experiments ON experiments.experiment_id = runs.experiment_id",
	).Where(
		"experiments.namespace_id = ?", namespace.ID,
	).Where(
		"runs.experiment_id IN ?", req.ExperimentIDs,
	).Where(
		"runs.lifecycle_stage IN ?", lifecyleStages,
	)

	// MaxResults
	// TODO if compatible with mlflow client, consider using same logic as in ExperimentSearch
	limit := int(req.MaxResults)
	if limit == 0 {
		limit = 1000
	}
	tx.Limit(limit)

	// PageToken
	var offset int
	if req.PageToken != "" {
		var token request.PageToken
		if err := json.NewDecoder(
			base64.NewDecoder(
				base64.StdEncoding,
				strings.NewReader(req.PageToken),
			),
		).Decode(&token); err != nil {
			return nil, 0, 0, api.NewInvalidParameterValueError("invalid page_token '%s': %s", req.PageToken, err)
		}
		offset = int(token.Offset)
	}
	tx.Offset(offset)

	// Filter
	if req.Filter != "" {
		for n, f := range filterAnd.Split(req.Filter, -1) {
			components := filterCond.FindStringSubmatch(f)
			if len(components) != 5 {
				return nil, 0, 0, api.NewInvalidParameterValueError("malformed filter '%s'", f)
			}

			entity := components[1]
			key := strings.Trim(components[2], "\"`")
			comparison := components[3]
			var value any = components[4]
			valueCol := "value"

			var kind any
			switch entity {
			case "", "attribute", "attributes", "attr", "run":
				switch key {
				case "start_time", "end_time":
					switch comparison {
					case GraterExpression, GraterOrEqualExpression, NotEqualExpression,
						EqualExpression, LessExpression, LessOrEqualExpression:
						v, err := strconv.Atoi(value.(string))
						if err != nil {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid numeric value '%s'", value)
						}
						value = v
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid numeric attribute comparison operator '%s'", comparison,
						)
					}
				case "run_name":
					key = "mlflow.runName"
					kind = &database.Tag{}
					fallthrough
				case "status", "user_id", "artifact_uri":
					switch strings.ToUpper(comparison) {
					case NotEqualExpression, EqualExpression, LikeExpression, ILikeExpression:
						if strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid string attribute comparison operator '%s'", comparison,
						)
					}
				case "run_id":
					key = "run_uuid"
					switch strings.ToUpper(comparison) {
					case NotEqualExpression, EqualExpression, LikeExpression, ILikeExpression:
						if strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					case InExpression, NotInExpression:
						if !strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid list definition '%s'", value)
						}
						var values []string
						for _, v := range filterInGroup.Split(value.(string)[1:len(value.(string))-1], -1) {
							values = append(values, strings.Trim(v, "'"))
						}
						value = values
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid string attribute comparison operator '%s'", comparison,
						)
					}
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError(
						`invalid attribute '%s'. `+
							`Valid values are ['run_name', 'start_time', 'end_time', 'status', 'user_id', 'artifact_uri', 'run_id']`,
						key,
					)
				}
			case "metric", "metrics":
				switch comparison {
				case GraterExpression, GraterOrEqualExpression,
					NotEqualExpression, EqualExpression, LessExpression, LessOrEqualExpression:
					v, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid numeric value '%s'", value)
					}
					value = v
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError(
						"invalid metric comparison operator '%s'", comparison,
					)
				}
				kind = &database.LatestMetric{}
			case "parameter", "parameters", "param", "params":
				switch strings.ToUpper(comparison) {
				case NotEqualExpression, EqualExpression:
					switch v := value.(type) {
					case string:
						if strings.HasPrefix(v, "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(v, `"'`)
						valueCol = "value_str"
					case int64, int32, int:
						valueCol = "value_int"
					case float64, float32:
						valueCol = "value_float"
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid value '%v' for comparison operator '%s'", v, comparison,
						)
					}
				case LikeExpression, ILikeExpression:
					switch v := value.(type) {
					case string:
						if strings.HasPrefix(v, "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(v, `"'`)
						valueCol = "value_str"
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid value '%v' for comparison operator '%s'", v, comparison,
						)
					}
				case GraterExpression, GraterOrEqualExpression, LessExpression, LessOrEqualExpression:
					switch v := value.(type) {
					case int64, int32, int:
						valueCol = "value_int"
					case float64, float32:
						valueCol = "value_float"
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError(
							"invalid value '%v' for comparison operator '%s'", v, comparison,
						)
					}
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError(
						"invalid param comparison operator '%s'", comparison,
					)
				}
				kind = &database.Param{}
			case "tag", "tags":
				switch strings.ToUpper(comparison) {
				case NotEqualExpression, EqualExpression, LikeExpression, ILikeExpression:
					if strings.HasPrefix(value.(string), "(") {
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError(
						"invalid tag comparison operator '%s'", comparison,
					)
				}
				kind = &database.Tag{}
			default:
				return nil, 0, 0, api.NewInvalidParameterValueError(
					"invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", entity,
				)
			}

			if kind == nil {
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == ILikeExpression {
					key = fmt.Sprintf("LOWER(runs.%s)", key)
					comparison = LikeExpression
					value = strings.ToLower(value.(string))
					tx.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
				} else {
					tx.Where(fmt.Sprintf("runs.%s %s ?", key, comparison), value)
				}
			} else {
				table := fmt.Sprintf("filter_%d", n)
				where := fmt.Sprintf("%s %s ?", valueCol, comparison)
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == ILikeExpression {
					where = fmt.Sprintf("LOWER(%s) LIKE ?", valueCol)
					value = strings.ToLower(fmt.Sprintf("%v", value))
				}
				tx.Joins(
					fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
					database.DB.Select("run_uuid", valueCol).Where("key = ?", key).Where(where, value).Model(kind),
				)
			}
		}
	}

	// OrderBy
	// TODO order numeric, nan, null?
	// TODO collation for strings on postgres?
	startTimeOrder := false
	for n, o := range req.OrderBy {
		components := runOrder.FindStringSubmatch(o)
		log.Debugf("Components: %#v", components)
		if len(components) < 3 {
			return nil, 0, 0, api.NewInvalidParameterValueError("invalid order_by clause '%s'", o)
		}

		column := strings.Trim(components[2], "`\"")

		var kind any
		switch components[1] {
		case "attribute":
			if column == "start_time" {
				startTimeOrder = true
			}
		case "metric":
			kind = &database.LatestMetric{}
		case "param":
			kind = &database.Param{}
		case "tag":
			kind = &database.Tag{}
		default:
			return nil, 0, 0, api.NewInvalidParameterValueError(
				"invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']",
				components[1],
			)
		}
		if kind != nil {
			table := fmt.Sprintf("order_%d", n)
			tx.Joins(
				fmt.Sprintf("LEFT OUTER JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
				database.DB.Select("run_uuid", "value").Where("key = ?", column).Model(kind),
			)
			column = fmt.Sprintf("%s.value", table)
		}
		tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: column,
			},
			Desc: len(components) == 4 && strings.ToUpper(components[3]) == "DESC",
		})
	}
	if !startTimeOrder {
		tx.Order("runs.start_time DESC")
	}
	tx.Order("runs.run_uuid")

	// Actual query
	var runs []models.Run
	tx.Preload("LatestMetrics").
		Preload("Params").
		Preload("Tags").
		Find(&runs)
	if tx.Error != nil {
		return nil, 0, 0, api.NewInternalError("unable to search runs: %s", tx.Error)
	}

	return runs, limit, offset, nil
}

// DeleteRun handles delete models.Run entity business logic.
func (s Service) DeleteRun(
	ctx context.Context, namespace *models.Namespace, req *request.DeleteRunRequest,
) error {
	if err := ValidateDeleteRunRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s'", req.RunID)
	}

	if err := s.runRepository.Archive(ctx, run); err != nil {
		return api.NewInternalError("unable to delete run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) RestoreRun(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.RestoreRunRequest,
) error {
	if err := ValidateRestoreRunRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s'", req.RunID)
	}

	run.DeletedTime = sql.NullInt64{Valid: false}
	run.LifecycleStage = models.LifecycleStageActive
	if err := s.runRepository.Update(ctx, run); err != nil {
		return api.NewInternalError("unable to restore run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) LogMetric(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.LogMetricRequest,
) error {
	if err := ValidateLogMetricRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s'", req.RunID)
	}

	metric, err := convertors.ConvertLogMetricRequestToDBModel(run.ID, req)
	if err != nil {
		return api.NewInvalidParameterValueError(err.Error())
	}
	if err := s.metricRepository.CreateBatch(ctx, run, 1, []models.Metric{*metric}); err != nil {
		return api.NewInternalError("unable to log metric '%s' for run '%s': %s", req.Key, req.GetRunID(), err)
	}

	return nil
}

func (s Service) LogParam(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.LogParamRequest,
) error {
	if err := ValidateLogParamRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDRunIDAndLifecycleStage(
		ctx, namespace.ID, req.RunID, models.LifecycleStageActive,
	)
	if err != nil {
		return api.NewInternalError("Unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("Run '%s' not found", req.RunID)
	}

	param := convertors.ConvertLogParamRequestToDBModel(run.ID, req)
	if err := s.paramRepository.CreateBatch(ctx, 1, []models.Param{*param}); err != nil {
		if errors.As(err, &repositories.ParamConflictError{}) {
			return api.NewInvalidParameterValueError("unable to insert params for run '%s': %s", run.ID, err)
		}
		return api.NewInternalError("unable to insert params for run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) SetRunTag(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.SetRunTagRequest,
) error {
	if err := ValidateSetRunTagRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDRunIDAndLifecycleStage(
		ctx, namespace.ID, req.RunID, models.LifecycleStageActive,
	)
	if err != nil {
		return api.NewInternalError("Unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("Run '%s' not found", req.RunID)
	}

	tag := convertors.ConvertSetRunTagRequestToDBModel(run.ID, req)
	if err := s.runRepository.SetRunTagsBatch(ctx, run, 1, []models.Tag{*tag}); err != nil {
		return api.NewInternalError("unable to insert tags for run '%s': %s", run.ID, err)
	}
	return nil
}

func (s Service) DeleteRunTag(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.DeleteRunTagRequest,
) error {
	if err := ValidateDeleteRunTagRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDRunIDAndLifecycleStage(
		ctx, namespace.ID, req.RunID, models.LifecycleStageActive,
	)
	if err != nil {
		return api.NewInternalError("Unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("Run '%s' not found", req.RunID)
	}

	tag, err := s.tagRepository.GetByRunIDAndKey(ctx, run.ID, req.Key)
	if err != nil {
		return api.NewInternalError("Unable to find tag '%s' for run '%s': %s", req.Key, req.RunID, err)
	}
	if tag == nil {
		return api.NewResourceDoesNotExistError("No tag with name: %s", req.Key)
	}

	if err := s.tagRepository.Delete(ctx, tag); err != nil {
		return api.NewInternalError("unable to delete tag '%s' for run '%s': %s", req.Key, req.RunID, err)
	}

	return nil
}

func (s Service) LogBatch(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.LogBatchRequest,
) error {
	if err := ValidateLogBatchRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDRunIDAndLifecycleStage(
		ctx, namespace.ID, req.RunID, models.LifecycleStageActive,
	)
	if err != nil {
		return api.NewInternalError("Unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("Run '%s' not found", req.RunID)
	}

	metrics, params, tags, err := convertors.ConvertLogBatchRequestToDBModel(run.ID, req)
	if err != nil {
		return api.NewInvalidParameterValueError(err.Error())
	}
	if err := s.paramRepository.CreateBatch(ctx, 100, params); err != nil {
		if errors.As(err, &repositories.ParamConflictError{}) {
			return api.NewInvalidParameterValueError("unable to insert params for run '%s': %s", run.ID, err)
		}
		return api.NewInternalError("unable to insert params for run '%s': %s", run.ID, err)
	}
	if err := s.metricRepository.CreateBatch(ctx, run, 100, metrics); err != nil {
		return api.NewInternalError("unable to insert metrics for run '%s': %s", run.ID, err)
	}
	if err := s.runRepository.SetRunTagsBatch(ctx, run, 100, tags); err != nil {
		return api.NewInternalError("unable to insert tags for run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) LogOutput(
	ctx context.Context,
	namespace *models.Namespace,
	req *request.LogOutputRequest,
) error {
	if err := ValidateLogOutputRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByNamespaceIDAndRunID(ctx, namespace.ID, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}
	if run == nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s'", req.RunID)
	}

	log := convertors.ConvertLogOutputRequestToDBModel(run.ID, req)
	if err := s.logRepository.Create(ctx, log); err != nil {
		return api.NewInternalError("unable to save log for run '%s'", req.RunID)
	}
	return nil
}

// LogArtifact creates new Run artifact.
func (s Service) LogArtifact(
	ctx context.Context, namespaceID uint, req *request.LogArtifactRequest,
) error {
	artifact := ConvertCreateRunArtifactRequestToModel(namespaceID, req)
	if err := s.artifactRepository.Create(ctx, artifact); err != nil {
		return api.NewInternalError("error creating run artifact: %s", err)
	}
	return nil
}
