package run

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/repositories"
)

var (
	filterAnd     = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond    = regexp.MustCompile(`^(?:(\w+)\.)?("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'[^']+'(?:,\s*)?)+\)|"[^"]+"|'[^']+'|[\w\.]+)$`)
	filterInGroup = regexp.MustCompile(`,\s*`)
	runOrder      = regexp.MustCompile(`^(attribute|metric|param|tag)s?\.("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)(?i:\s+(ASC|DESC))?$`)
)

// Service provides service layer to work with `run` business logic.
type Service struct {
	tagRepository        repositories.TagRepositoryProvider
	runRepository        repositories.RunRepositoryProvider
	paramRepository      repositories.ParamRepositoryProvider
	metricRepository     repositories.MetricRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	tagRepository repositories.TagRepositoryProvider,
	runRepository repositories.RunRepositoryProvider,
	paramRepository repositories.ParamRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		tagRepository:        tagRepository,
		runRepository:        runRepository,
		paramRepository:      paramRepository,
		metricRepository:     metricRepository,
		experimentRepository: experimentRepository,
	}
}

func (s Service) CreateRun(ctx context.Context, req *request.CreateRunRequest) (*models.Run, error) {
	experimentID, err := strconv.ParseInt(req.ExperimentID, 10, 32)
	if err != nil {
		return nil, api.NewBadRequestError("unable to parse experiment id '%s': %s", req.ExperimentID, err)
	}

	experiment, err := s.experimentRepository.GetByID(ctx, int32(experimentID))
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find experiment '%v': %s", experiment, err)
	}

	run := convertors.ConvertCreateRunRequestToDBModel(experiment, req)
	if err := s.runRepository.Create(ctx, run); err != nil {
		return nil, api.NewInternalError("error inserting run '%s': %s", run.ID, err)
	}

	return run, nil
}

func (s Service) UpdateRun(ctx context.Context, req *request.UpdateRunRequest) (*models.Run, error) {
	if err := ValidateUpdateRunRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByID(ctx, req.GetRunID())
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}

	run = convertors.ConvertUpdateRunRequestToDBModel(run, req)
	if err := s.runRepository.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := s.runRepository.UpdateWithTransaction(ctx, tx, run); err != nil {
			return err
		}
		if req.Name != "" {
			// TODO:DSuhinin - move "mlflow.runName" to be a constant somewhere.
			// Also Im not fully sure that this is right place to keep this logic here.
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

func (s Service) GetRun(ctx context.Context, req *request.GetRunRequest) (*models.Run, error) {
	if err := ValidateGetRunRequest(req); err != nil {
		return nil, err
	}

	run, err := s.runRepository.GetByID(ctx, req.GetRunID())
	if err != nil {
		return nil, api.NewResourceDoesNotExistError("unable to find run '%s': %s", run.ID, err)
	}

	return run, nil
}

func (s Service) SearchRuns(ctx context.Context, req *request.SearchRunsRequest) ([]models.Run, int, int, error) {
	if err := ValidateSearchRunsRequest(req); err != nil {
		return nil, 0, 0, err
	}

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
	tx := database.DB.Where(
		"experiment_id IN ?", req.ExperimentIDs,
	).Where(
		"lifecycle_stage IN ?", lifecyleStages,
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

			var kind any
			switch entity {
			case "", "attribute", "attributes", "attr", "run":
				switch key {
				case "start_time", "end_time":
					switch comparison {
					case ">", ">=", "!=", "=", "<", "<=":
						v, err := strconv.Atoi(value.(string))
						if err != nil {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid numeric value '%s'", value)
						}
						value = v
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid numeric attribute comparison operator '%s'", comparison)
					}
				case "run_name":
					key = "mlflow.runName"
					kind = &database.Tag{}
					fallthrough
				case "status", "user_id", "artifact_uri":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string attribute comparison operator '%s'", comparison)
					}
				case "run_id":
					key = "run_uuid"
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					case "IN", "NOT IN":
						if !strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid list definition '%s'", value)
						}
						var values []string
						for _, v := range filterInGroup.Split(value.(string)[1:len(value.(string))-1], -1) {
							values = append(values, strings.Trim(v, "'"))
						}
						value = values
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string attribute comparison operator '%s'", comparison)
					}
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError("invalid attribute '%s'. Valid values are ['run_name', 'start_time', 'end_time', 'status', 'user_id', 'artifact_uri', 'run_id']", key)
				}
			case "metric", "metrics":
				switch comparison {
				case ">", ">=", "!=", "=", "<", "<=":
					v, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid numeric value '%s'", value)
					}
					value = v
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError("invalid metric comparison operator '%s'", comparison)
				}
				kind = &database.LatestMetric{}
			case "parameter", "parameters", "param", "params":
				switch strings.ToUpper(comparison) {
				case "!=", "=", "LIKE", "ILIKE":
					if strings.HasPrefix(value.(string), "(") {
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError("invalid param comparison operator '%s'", comparison)
				}
				kind = &database.Param{}
			case "tag", "tags":
				switch strings.ToUpper(comparison) {
				case "!=", "=", "LIKE", "ILIKE":
					if strings.HasPrefix(value.(string), "(") {
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError("invalid tag comparison operator '%s'", comparison)
				}
				kind = &database.Tag{}
			default:
				return nil, 0, 0, api.NewInvalidParameterValueError("invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", entity)
			}

			if kind == nil {
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
					key = fmt.Sprintf("LOWER(%s)", key)
					comparison = "LIKE"
					value = strings.ToLower(value.(string))
				}
				tx.Where(fmt.Sprintf("runs.%s %s ?", key, comparison), value)
			} else {
				table := fmt.Sprintf("filter_%d", n)
				where := fmt.Sprintf("value %s ?", comparison)
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
					where = "LOWER(value) LIKE ?"
					value = strings.ToLower(value.(string))
				}
				tx.Joins(
					fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
					database.DB.Select("run_uuid", "value").Where("key = ?", key).Where(where, value).Model(kind),
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
			return nil, 0, 0, api.NewInvalidParameterValueError("invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", components[1])
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
func (s Service) DeleteRun(ctx context.Context, req *request.DeleteRunRequest) error {
	if err := ValidateDeleteRunRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}

	if err := s.runRepository.Delete(ctx, run); err != nil {
		return api.NewInternalError("unable to delete run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) RestoreRun(ctx context.Context, req *request.RestoreRunRequest) error {
	if err := ValidateRestoreRunRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}

	if err := s.runRepository.Restore(ctx, run); err != nil {
		return api.NewInternalError("unable to restore run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) LogMetric(ctx context.Context, req *request.LogMetricRequest) error {
	if err := ValidateLogMetricRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find run '%s': %s", req.RunID, err)
	}

	if err := logMetrics(run.ID, []request.MetricPartialRequest{{
		Key:       req.Key,
		Step:      req.Step,
		Value:     req.Value,
		Timestamp: req.Timestamp,
	}}); err != nil {
		return api.NewInternalError("unable to log metric '%s' for run '%s': %s", req.Key, req.GetRunID(), err)
	}

	return nil
}

func (s Service) LogParam(ctx context.Context, req *request.LogParamRequest) error {
	if err := ValidateLogParamRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil || !run.IsLifecycleStageActive() {
		return api.NewResourceDoesNotExistError("unable to find active run '%s': %s", req.RunID, err)
	}

	param := convertors.ConvertLogParamRequestToDBModel(run.ID, req)
	if err := s.paramRepository.Create(ctx, param); err != nil {
		return api.NewInternalError("unable to insert params for run '%s': %s", run.ID, err)
	}

	return nil
}

func (s Service) SetRunTag(ctx context.Context, req *request.SetRunTagRequest) error {
	if err := ValidateSetRunTagRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil || !run.IsLifecycleStageActive() {
		return api.NewResourceDoesNotExistError("unable to find active run '%s': %s", req.RunID, err)
	}

	tag := convertors.ConvertSetRunTagRequestToDBModel(run.ID, req)
	if err := s.runRepository.SetRunTagsBatch(ctx, 1, run, []models.Tag{*tag}); err != nil {
		return api.NewInternalError("unable to insert tags for run '%s': %s", run.ID, err)
	}
	return nil
}

func (s Service) DeleteRunTag(ctx context.Context, req *request.DeleteRunTagRequest) error {
	if err := ValidateDeleteRunTagRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil || !run.IsLifecycleStageActive() {
		return api.NewResourceDoesNotExistError("Unable to find active run '%s': %s", req.RunID, err)
	}

	tag, err := s.tagRepository.GetByRunIDAndKey(ctx, run.ID, req.Key)
	if err != nil {
		return api.NewResourceDoesNotExistError("Unable to find tag '%s' for run '%s': %s", req.Key, req.RunID, err)
	}

	if err := s.tagRepository.Delete(ctx, tag); err != nil {
		return api.NewInternalError("unable to delete tag '%s' for run '%s': %s", req.Key, req.RunID, err)
	}

	return nil
}

func (s Service) LogBatch(ctx context.Context, req *request.LogBatchRequest) error {
	if err := ValidateLogBatchRequest(req); err != nil {
		return err
	}

	run, err := s.runRepository.GetByID(ctx, req.RunID)
	if err != nil || !run.IsLifecycleStageActive() {
		return api.NewResourceDoesNotExistError("unable to find active run '%s': %s", req.RunID, err)
	}

	params, tags := convertors.ConvertLogBatchRequestToDBModel(run.ID, req)
	if err := s.paramRepository.CreateBatch(ctx, 100, params); err != nil {
		return api.NewInternalError("unable to insert params for run '%s': %s", run.ID, err)
	}

	if err := logMetrics(req.RunID, req.Metrics); err != nil {
		return err
	}

	if err := s.runRepository.SetRunTagsBatch(ctx, 100, run, tags); err != nil {
		return api.NewInternalError("unable to insert tags for run '%s': %s", run.ID, err)
	}

	return nil
}

func logMetrics(id string, metrics []request.MetricPartialRequest) error {
	if len(metrics) == 0 {
		return nil
	}

	if tx := database.DB.Select(
		"run_uuid",
	).Where(
		"lifecycle_stage = ?", database.LifecycleStageActive,
	).First(
		&database.Run{ID: id},
	); tx.Error != nil {
		return api.NewResourceDoesNotExistError("unable to find active run '%s': %s", id, tx.Error)
	}

	lastIters := make(map[string]int64)
	for _, m := range metrics {
		lastIters[m.Key] = -1
	}
	keys := make([]string, 0, len(lastIters))
	for k := range lastIters {
		keys = append(keys, k)
	}

	if err := func() error {
		rows, err := database.DB.Table("latest_metrics").Select("key", "last_iter").Where("run_uuid = ?", id).Where("key IN ?", keys).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var key string
			var iter int64
			if err := rows.Scan(&key, &iter); err != nil {
				return err
			}
			lastIters[key] = iter
		}

		return nil
	}(); err != nil {
		return api.NewInternalError("unable to get latest metric iters for run '%s': %s", id, err)
	}

	dbMetrics := make([]database.Metric, len(metrics))
	latestMetrics := make(map[string]database.LatestMetric)
	for n, metric := range metrics {
		m := database.Metric{
			RunID:     id,
			Key:       metric.Key,
			Timestamp: metric.Timestamp,
			Step:      metric.Step,
			Iter:      lastIters[metric.Key] + 1,
		}
		if v, ok := metric.Value.(float64); ok {
			m.Value = v
		} else if v, ok := metric.Value.(string); ok {
			switch v {
			case "NaN":
				m.Value = 0
				m.IsNan = true
			case "Infinity":
				m.Value = math.MaxFloat64
				// m.Value = math.Inf(1)
			case "-Infinity":
				m.Value = -math.MaxFloat64
				// m.Value = math.Inf(-1)
			default:
				return api.NewInvalidParameterValueError("invalid metric value '%s'", v)
			}
		} else {
			return api.NewInvalidParameterValueError("invalid metric value '%s'", v)
		}
		dbMetrics[n] = m

		lastIters[metric.Key] = m.Iter

		lm, ok := latestMetrics[m.Key]
		if !ok ||
			m.Step > lm.Step ||
			(m.Step == lm.Step && m.Timestamp > lm.Timestamp) ||
			(m.Step == lm.Step && m.Timestamp == lm.Timestamp && m.Value > lm.Value) {
			latestMetrics[m.Key] = database.LatestMetric{
				RunID:     m.RunID,
				Key:       m.Key,
				Value:     m.Value,
				Timestamp: m.Timestamp,
				Step:      m.Step,
				IsNan:     m.IsNan,
				LastIter:  m.Iter,
			}
		}
	}

	if tx := database.DB.CreateInBatches(&dbMetrics, 100); tx.Error != nil {
		return api.NewInternalError("unable to insert metrics for run '%s': %s", id, tx.Error)
	}

	// TODO update latest metrics in the background?

	var currentLatestMetrics []database.LatestMetric
	if tx := database.DB.Where("run_uuid = ?", id).Where("key IN ?", keys).Find(&currentLatestMetrics); tx.Error != nil {
		return api.NewInternalError("unable to get latest metrics for run '%s': %s", id, tx.Error)
	}

	currentLatestMetricsMap := make(map[string]database.LatestMetric, len(currentLatestMetrics))
	for _, m := range currentLatestMetrics {
		currentLatestMetricsMap[m.Key] = m
	}

	updatedLatestMetrics := make([]database.LatestMetric, 0, len(latestMetrics))
	for k, m := range latestMetrics {
		lm, ok := currentLatestMetricsMap[k]
		if !ok ||
			m.Step > lm.Step ||
			(m.Step == lm.Step && m.Timestamp > lm.Timestamp) ||
			(m.Step == lm.Step && m.Timestamp == lm.Timestamp && m.Value > lm.Value) {
			updatedLatestMetrics = append(updatedLatestMetrics, m)
		} else {
			lm.LastIter = lastIters[k]
			updatedLatestMetrics = append(updatedLatestMetrics, lm)
		}
	}

	if len(updatedLatestMetrics) > 0 {
		if tx := database.DB.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&updatedLatestMetrics); tx.Error != nil {
			return api.NewInternalError("unable to update latest metrics for run '%s': %s", id, tx.Error)
		}
	}

	return nil
}
