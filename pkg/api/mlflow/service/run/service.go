package run

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/service/experiment"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrack/pkg/database"
)

var (
	filterAnd     = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond    = regexp.MustCompile(`^(?:(\w+)\.)?("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'[^']+'(?:,\s*)?)+\)|"[^"]+"|'[^']+'|[\w\.]+)$`)
	filterInGroup = regexp.MustCompile(`,\s*`)
	runOrder      = regexp.MustCompile(`^(attribute|metric|param|tag)s?\.("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)(?i:\s+(ASC|DESC))?$`)
)

func CreateRun(c *fiber.Ctx) error {
	var req request.CreateRunRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("CreateRun request: %#v", &req)

	ex, err := strconv.ParseInt(req.ExperimentID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ExperimentID, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID: &ex32,
	}
	if tx := database.DB.Select("artifact_location").First(&exp); tx.Error != nil {
		return api.NewResourceDoesNotExistError("Unable to find experiment '%d': %s", ex, tx.Error)
	}

	run := database.Run{
		ID:           database.NewUUID(),
		Name:         req.Name,
		ExperimentID: *exp.ID,
		UserID:       req.UserID,
		Status:       database.StatusRunning,
		StartTime: sql.NullInt64{
			Int64: req.StartTime,
			Valid: true,
		},
		LifecycleStage: database.LifecycleStageActive,
		Tags:           make([]database.Tag, len(req.Tags)),
	}

	run.ArtifactURI = fmt.Sprintf("%s/%s/artifacts", exp.ArtifactLocation, run.ID)

	for n, tag := range req.Tags {
		switch tag.Key {
		case "mlflow.user":
			if run.UserID == "" {
				run.UserID = tag.Value
			}
		case "mlflow.source.name":
			run.SourceName = tag.Value
		case "mlflow.source.type":
			run.SourceType = tag.Value
		case "mlflow.runName":
			if run.Name == "" {
				run.Name = tag.Value
			}
		}
		run.Tags[n] = database.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	if run.Name == "" {
		run.Name = experiment.GenerateRandomName()
		run.Tags = append(run.Tags, database.Tag{
			Key:   "mlflow.runName",
			Value: run.Name,
		})
	}

	if run.SourceType == "" {
		run.SourceType = "UNKNOWN"
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
				return err
			}
		}
		return tx.Create(&run).Error
	}); err != nil {
		return api.NewInternalError("Error inserting run '%s': %s", run.ID, err)
	}

	resp := response.NewCreateRunResponse(&run)

	log.Debugf("CreateRun response: %#v", resp)

	return c.JSON(resp)
}

func UpdateRun(c *fiber.Ctx) error {
	var req request.UpdateRunRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("UpdateRun request: %#v", &req)

	if err := ValidateUpdateRunRequest(&req); err != nil {
		return err
	}

	run := database.Run{
		ID: req.RunID,
	}
	if run.ID == "" {
		run.ID = req.RunUUID
	}
	if err := database.DB.First(&run).Error; err != nil {
		return api.NewInvalidParameterValueError("Unable to find run '%s': %s", run.ID, err)
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&run).Updates(database.Run{
			Name:   req.Name,
			Status: database.Status(req.Status),
			EndTime: sql.NullInt64{
				Int64: req.EndTime,
				Valid: true,
			},
		}).Error; err != nil {
			return err
		}

		if req.Name != "" {
			if err := tx.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create([]database.Tag{{
				Key:   "mlflow.runName",
				Value: req.Name,
				RunID: run.ID,
			}}).Error; err != nil {
				return err
			}
		}
		return nil
	}).Error; err != nil {
		return api.NewInternalError("Unable to update run '%s': %v", run.ID, err)
	}

	resp := response.NewUpdateRunResponse(&run)

	log.Debugf("UpdateRun response: %#v", resp)

	return c.JSON(resp)
}

func GetRun(c *fiber.Ctx) error {
	req := request.GetRunRequest{}
	if err := c.QueryParser(&req); err != nil {
		return api.NewBadRequestError(err.Error())
	}

	log.Debugf("GetRun request: run_id=%q, run_uuid=%q", req.RunID, req.RunUUID)

	if err := ValidateGetRunRequest(&req); err != nil {
		return err
	}

	run := database.Run{
		ID: req.GetRunID(),
	}
	if err := database.DB.Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload("Tags").First(&run).Error; err != nil {
		return api.NewResourceDoesNotExistError("Unable to find run '%s': %s", run.ID, err)
	}

	resp := &response.GetRunResponse{
		Run: modelRunToAPI(run),
	}

	log.Debugf("GetRun response: %#v", resp)

	return c.JSON(resp)
}

func SearchRuns(c *fiber.Ctx) error {
	var req request.SearchRunsRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("SearchRuns request: %#v", req)

	if err := ValidateSearchRunsRequest(&req); err != nil {
		return err
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
			return api.NewInvalidParameterValueError("Invalid page_token '%s': %s", req.PageToken, err)

		}
		offset = int(token.Offset)
	}
	tx.Offset(offset)

	// Filter
	if req.Filter != "" {
		for n, f := range filterAnd.Split(req.Filter, -1) {
			components := filterCond.FindStringSubmatch(f)
			if len(components) != 5 {
				return api.NewInvalidParameterValueError("Malformed filter '%s'", f)
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
							return api.NewInvalidParameterValueError("Invalid numeric value '%s'", value)
						}
						value = v
					default:
						return api.NewInvalidParameterValueError("Invalid numeric attribute comparison operator '%s'", comparison)
					}
				case "run_name":
					key = "mlflow.runName"
					kind = &database.Tag{}
					fallthrough
				case "status", "user_id", "artifact_uri":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return api.NewInvalidParameterValueError("Invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					default:
						return api.NewInvalidParameterValueError("Invalid string attribute comparison operator '%s'", comparison)
					}
				case "run_id":
					key = "run_uuid"
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return api.NewInvalidParameterValueError("Invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					case "IN", "NOT IN":
						if !strings.HasPrefix(value.(string), "(") {
							return api.NewInvalidParameterValueError("Invalid list definition '%s'", value)
						}
						var values []string
						for _, v := range filterInGroup.Split(value.(string)[1:len(value.(string))-1], -1) {
							values = append(values, strings.Trim(v, "'"))
						}
						value = values
					default:
						return api.NewInvalidParameterValueError("Invalid string attribute comparison operator '%s'", comparison)
					}
				default:
					return api.NewInvalidParameterValueError("Invalid attribute '%s'. Valid values are ['run_name', 'start_time', 'end_time', 'status', 'user_id', 'artifact_uri', 'run_id']", key)
				}
			case "metric", "metrics":
				switch comparison {
				case ">", ">=", "!=", "=", "<", "<=":
					v, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						return api.NewInvalidParameterValueError("Invalid numeric value '%s'", value)
					}
					value = v
				default:
					return api.NewInvalidParameterValueError("Invalid metric comparison operator '%s'", comparison)
				}
				kind = &database.LatestMetric{}
			case "parameter", "parameters", "param", "params":
				switch strings.ToUpper(comparison) {
				case "!=", "=", "LIKE", "ILIKE":
					if strings.HasPrefix(value.(string), "(") {
						return api.NewInvalidParameterValueError("Invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return api.NewInvalidParameterValueError("Invalid param comparison operator '%s'", comparison)
				}
				kind = &database.Param{}
			case "tag", "tags":
				switch strings.ToUpper(comparison) {
				case "!=", "=", "LIKE", "ILIKE":
					if strings.HasPrefix(value.(string), "(") {
						return api.NewInvalidParameterValueError("Invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return api.NewInvalidParameterValueError("Invalid tag comparison operator '%s'", comparison)
				}
				kind = &database.Tag{}
			default:
				return api.NewInvalidParameterValueError("Invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", entity)
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
			return api.NewInvalidParameterValueError("Invalid order_by clause '%s'", o)
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
			return api.NewInvalidParameterValueError("Invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", components[1])
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
	var runs []database.Run
	tx.Preload("LatestMetrics").
		Preload("Params").
		Preload("Tags").
		Find(&runs)
	if tx.Error != nil {
		return api.NewInternalError("Unable to search runs: %s", tx.Error)
	}

	resp := &response.SearchRunsResponse{
		Runs: make([]response.RunPartialResponse, len(runs)),
	}
	for n, r := range runs {
		resp.Runs[n] = modelRunToAPI(r)
	}

	// NextPageToken
	if len(runs) == limit {
		var token strings.Builder
		b64 := base64.NewEncoder(base64.StdEncoding, &token)
		if err := json.NewEncoder(b64).Encode(request.PageToken{
			Offset: int32(offset + limit),
		}); err != nil {
			return api.NewInternalError("Unable to build next_page_token: %s", err)
		}
		b64.Close()
		resp.NextPageToken = token.String()
	}

	log.Debugf("SearchRuns response: %#v", resp)

	return c.JSON(resp)
}

func DeleteRun(c *fiber.Ctx) error {
	var req request.DeleteRunRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("DeleteRun request: %#v", req)

	if err := ValidateDeleteRunRequest(&req); err != nil {
		return err
	}

	run := database.Run{ID: req.RunID}
	if tx := database.DB.Select("lifecycle_stage").First(&run); tx.Error != nil {
		return api.NewInvalidParameterValueError("Unable to find run '%s': %s", run.ID, tx.Error)
	}

	if err := database.DB.Model(&run).Updates(database.Run{
		LifecycleStage: database.LifecycleStageDeleted,
		DeletedTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
	}).Error; err != nil {
		return api.NewInternalError("Unable to update run '%s': %s", run.ID, err)
	}

	return c.JSON(fiber.Map{})
}

func RestoreRun(c *fiber.Ctx) error {
	var req request.RestoreRunRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("RestoreRun request: %#v", req)

	if err := ValidateRestoreRunRequest(&req); err != nil {
		return err
	}

	run := database.Run{ID: req.RunID}
	if err := database.DB.Select("lifecycle_stage").First(&run).Error; err != nil {
		return api.NewResourceDoesNotExistError("Unable to find run '%s': %s", run.ID, err)
	}

	// Use UpdateColumns so we can reset DeletedTime to null
	if err := database.DB.Model(&run).UpdateColumns(map[string]any{
		"DeletedTime":    sql.NullInt64{},
		"LifecycleStage": database.LifecycleStageActive,
	}).Error; err != nil {
		return api.NewInternalError("Unable to update run '%s': %s", run.ID, err)
	}

	return c.JSON(fiber.Map{})
}

func LogMetric(c *fiber.Ctx) error {
	var req request.LogMetricRequest
	if err := c.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("LogMetric request: %#v", req)

	if err := ValidateLogMetricRequest(&req); err != nil {
		return err
	}

	if err := logMetrics(req.GetRunID(), []request.MetricPartialRequest{{
		Key:       req.Key,
		Step:      req.Step,
		Value:     req.Value,
		Timestamp: req.Timestamp,
	}}); err != nil {
		return api.NewInternalError("Unable to log metric '%s' for run '%s': %s", req.Key, req.GetRunID(), err)
	}

	return c.JSON(fiber.Map{})
}

func LogParam(c *fiber.Ctx) error {
	var req request.LogParamRequest
	if err := c.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("LogParam request: %#v", req)

	if err := ValidateLogParamRequest(&req); err != nil {
		return err
	}

	if err := logParams(req.GetRunID(), []request.ParamPartialRequest{{Key: req.Key, Value: req.Value}}); err != nil {
		return api.NewInternalError("Unable to log param '%s' for run '%s': %s", req.Key, req.GetRunID(), err)
	}

	return c.JSON(fiber.Map{})
}

func SetRunTag(c *fiber.Ctx) error {
	var req request.SetRunTagRequest
	if err := c.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("SetRunTag request: %#v", req)

	if err := ValidateSetRunTagRequest(&req); err != nil {
		return err
	}

	if err := setRunTags(req.GetRunID(), []request.TagPartialRequest{{Key: req.Key, Value: req.Value}}); err != nil {
		return api.NewInternalError("Unable to set tag '%s' for run '%s': %s", req.Key, req.GetRunID(), err)
	}

	return c.JSON(fiber.Map{})
}

func DeleteRunTag(c *fiber.Ctx) error {
	var req request.DeleteRunTagRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("DeleteRunTag request: %#v", req)

	if err := ValidateDeleteRunTagRequest(&req); err != nil {
		return err
	}
	if err := database.DB.Select(
		"run_uuid",
	).Where(
		"lifecycle_stage = ?", database.LifecycleStageActive,
	).First(
		&database.Run{ID: req.ID},
	).Error; err != nil {
		return api.NewResourceDoesNotExistError("Unable to find active run '%s': %s", req.ID, err)
	}

	if err := database.DB.First(&database.Tag{RunID: req.ID, Key: req.Key}).Error; err != nil {
		return api.NewResourceDoesNotExistError("Unable to find tag '%s' for run '%s': %s", req.Key, req.ID, err)
	}

	if tx := database.DB.Delete(&database.Tag{
		RunID: req.ID,
		Key:   req.Key,
	}); tx.Error != nil {
		return api.NewInternalError("Unable to delete tag '%s' for run '%s': %s", req.Key, req.ID, tx.Error)
	}

	return c.JSON(fiber.Map{})
}

func LogBatch(c *fiber.Ctx) error {
	var req request.LogBatchRequest
	if err := c.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("LogBatch request: %#v", req)

	if err := ValidateLogBatchRequest(&req); err != nil {
		return err
	}

	if err := logParams(req.ID, req.Params); err != nil {
		return err
	}

	if err := logMetrics(req.ID, req.Metrics); err != nil {
		return err
	}

	if err := setRunTags(req.ID, req.Tags); err != nil {
		return err
	}

	return c.JSON(fiber.Map{})
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
		return api.NewResourceDoesNotExistError("Unable to find active run '%s': %s", id, tx.Error)
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
		return api.NewInternalError("Unable to get latest metric iters for run '%s': %s", id, err)
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
				return api.NewInvalidParameterValueError("Invalid metric value '%s'", v)
			}
		} else {
			return api.NewInvalidParameterValueError("Invalid metric value '%s'", v)
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
		return api.NewInternalError("Unable to insert metrics for run '%s': %s", id, tx.Error)
	}

	// TODO update latest metrics in the background?

	var currentLatestMetrics []database.LatestMetric
	if tx := database.DB.Where("run_uuid = ?", id).Where("key IN ?", keys).Find(&currentLatestMetrics); tx.Error != nil {
		return api.NewInternalError("Unable to get latest metrics for run '%s': %s", id, tx.Error)
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
			return api.NewInternalError("Unable to update latest metrics for run '%s': %s", id, tx.Error)
		}
	}

	return nil
}

func logParams(id string, params []request.ParamPartialRequest) error {
	if len(params) == 0 {
		return nil
	}

	if tx := database.DB.Select("run_uuid").Where("lifecycle_stage = ?", database.LifecycleStageActive).First(&database.Run{ID: id}); tx.Error != nil {
		return api.NewResourceDoesNotExistError("Unable to find active run '%s': %s", id, tx.Error)
	}

	dbParams := make([]database.Param, len(params))
	for n, p := range params {
		dbParams[n] = database.Param{
			Key:   p.Key,
			Value: p.Value,
			RunID: id,
		}
	}

	if tx := database.DB.CreateInBatches(&dbParams, 100); tx.Error != nil {
		return api.NewInternalError("Unable to insert params for run '%s': %s", id, tx.Error)
	}

	return nil
}

func setRunTags(id string, tags []request.TagPartialRequest) error {
	if len(tags) == 0 {
		return nil
	}

	run := database.Run{ID: id}
	if tx := database.DB.Select("run_uuid", "name", "user_id").Where("lifecycle_stage = ?", database.LifecycleStageActive).First(&run); tx.Error != nil {
		return api.NewResourceDoesNotExistError("Unable to find active run '%s': %s", id, tx.Error)
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		dbTags := make([]database.Tag, len(tags))
		for n, t := range tags {
			dbTags[n] = database.Tag{
				Key:   t.Key,
				Value: t.Value,
				RunID: id,
			}
			switch t.Key {
			case "mlflow.runName":
				if run.Name != t.Value {
					if err := tx.Model(&run).UpdateColumn("name", t.Value).Error; err != nil {
						return err
					}
				}
			case "mlflow.user":
				if run.UserID != t.Value {
					if err := tx.Model(&run).UpdateColumn("user_id", t.Value).Error; err != nil {
						return err
					}
				}
			}
		}
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).CreateInBatches(&dbTags, 100).Error; err != nil {
			return err
		}
		return nil
	}).Error; err != nil {
		return api.NewInternalError("Unable to insert tags for run '%s': %s", id, err())
	}

	return nil
}

func modelRunToAPI(r database.Run) response.RunPartialResponse {
	metrics := make([]response.RunMetricPartialResponse, len(r.LatestMetrics))
	for n, m := range r.LatestMetrics {
		metrics[n] = response.RunMetricPartialResponse{
			Key:       m.Key,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Step:      m.Step,
		}
		if m.IsNan {
			metrics[n].Value = "NaN"
		}
	}

	params := make([]response.RunParamPartialResponse, len(r.Params))
	for n, p := range r.Params {
		params[n] = response.RunParamPartialResponse{
			Key:   p.Key,
			Value: p.Value,
		}
	}

	tags := make([]response.RunTagPartialResponse, len(r.Tags))
	for n, t := range r.Tags {
		tags[n] = response.RunTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
		switch t.Key {
		case "mlflow.runName":
			r.Name = t.Value
		case "mlflow.user":
			r.UserID = t.Value
		}
	}

	return response.RunPartialResponse{
		Info: response.RunInfoPartialResponse{
			ID:             r.ID,
			UUID:           r.ID,
			Name:           r.Name,
			ExperimentID:   fmt.Sprint(r.ExperimentID),
			UserID:         r.UserID,
			Status:         string(r.Status),
			StartTime:      r.StartTime.Int64,
			EndTime:        r.EndTime.Int64,
			ArtifactURI:    r.ArtifactURI,
			LifecycleStage: string(r.LifecycleStage),
		},
		Data: response.RunDataPartialResponse{
			Metrics: metrics,
			Params:  params,
			Tags:    tags,
		},
	}
}
