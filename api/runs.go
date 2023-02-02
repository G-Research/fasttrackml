package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fasttrack/model"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	filterAnd     *regexp.Regexp = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond    *regexp.Regexp = regexp.MustCompile(`^(?:(\w+)\.)?("[\w\.\- ]+"|` + "`" + `[\w\.\- ]+` + "`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'\w{32}'(?:,\s*)?)+\)|"[\w\.\- ]+"|'[\w\.\- %]+'|[\w\.]+)$`)
	filterInGroup *regexp.Regexp = regexp.MustCompile(`,\s*`)
	runOrder      *regexp.Regexp = regexp.MustCompile(`^(attribute|metric|param|tag)s?\.("[\w\.\- ]+"|` + "`" + `[\w\.\- ]+` + "`" + `|[\w\.]+)(?i:\s+(ASC|DESC))?$`)
)

func RunCreate(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunCreate request: %#v", &req)

		ex, err := strconv.ParseInt(req.ExperimentID, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", req.ExperimentID, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID: &ex32,
		}
		if tx := db.Select("artifact_location").First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", ex, tx.Error)
		}

		run := model.Run{
			ID:           model.NewUUID(),
			Name:         req.Name,
			ExperimentID: *exp.ID,
			UserID:       req.UserID,
			Status:       model.StatusRunning,
			StartTime: sql.NullInt64{
				Int64: req.StartTime,
				Valid: true,
			},
			LifecycleStage: model.LifecycleStageActive,
			Tags:           make([]model.Tag, len(req.Tags)),
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
			run.Tags[n] = model.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			}
		}

		if run.Name == "" {
			run.Name = GenerateRandomName()
			run.Tags = append(run.Tags, model.Tag{
				Key:   "mlflow.runName",
				Value: run.Name,
			})
		}

		if run.SourceType == "" {
			run.SourceType = "UNKNOWN"
		}

		if tx := db.Create(&run); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Error inserting run '%s': %s", run.ID, tx.Error)
		}

		resp := &RunCreateResponse{
			Run: Run{
				Info: RunInfo{
					ID:             run.ID,
					UUID:           run.ID,
					Name:           run.Name,
					ExperimentID:   fmt.Sprint(run.ExperimentID),
					UserID:         run.UserID,
					Status:         RunStatus(run.Status),
					StartTime:      run.StartTime.Int64,
					ArtifactURI:    run.ArtifactURI,
					LifecycleStage: LifecycleStage(run.LifecycleStage),
				},
				Data: RunData{
					Tags: make([]RunTag, len(run.Tags)),
				},
			},
		}
		for n, tag := range run.Tags {
			resp.Run.Data.Tags[n] = RunTag{
				Key:   tag.Key,
				Value: tag.Value,
			}
		}

		log.Debugf("RunCreate response: %#v", resp)

		return resp
	},
		http.MethodPost,
	))
}

func RunUpdate(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunUpdate request: %#v", &req)

		if req.ID == "" && req.UUID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		run := model.Run{
			ID: req.ID,
		}
		if run.ID == "" {
			run.ID = req.UUID
		}
		if tx := db.First(&run); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find run '%s': %s", run.ID, tx.Error)
		}

		tx := db.Begin()
		tx.Model(&run).Updates(model.Run{
			Name:   req.Name,
			Status: model.Status(req.Status),
			EndTime: sql.NullInt64{
				Int64: req.EndTime,
				Valid: true,
			},
		})

		if req.Name != "" {
			tx.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create([]model.Tag{{
				Key:   "mlflow.runName",
				Value: req.Name,
				RunID: run.ID,
			}})
		}

		tx.Commit()
		if tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to update run '%s': %s", run.ID, tx.Error)
		}

		// TODO grab name and user from tags?
		resp := &RunUpdateResponse{
			RunInfo: RunInfo{
				ID:             run.ID,
				UUID:           run.ID,
				Name:           run.Name,
				ExperimentID:   fmt.Sprint(run.ExperimentID),
				UserID:         run.UserID,
				Status:         RunStatus(run.Status),
				StartTime:      run.StartTime.Int64,
				EndTime:        run.EndTime.Int64,
				ArtifactURI:    run.ArtifactURI,
				LifecycleStage: LifecycleStage(run.LifecycleStage),
			},
		}

		log.Debugf("RunUpdate response: %#v", resp)

		return resp
	},
		http.MethodPost,
	))
}

func RunGet(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		id := r.URL.Query().Get("run_id")
		if id == "" {
			id = r.URL.Query().Get("run_uuid")
		}

		log.Debugf("RunGet request: run_id='%s'", id)

		if id == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		run := model.Run{
			ID: id,
		}
		if tx := db.Preload("LatestMetrics").Preload("Params").Preload("Tags").First(&run); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find run '%s': %s", run.ID, tx.Error)
		}

		resp := &RunGetResponse{
			Run: modelRunToAPI(run),
		}

		log.Debugf("RunGet response: %#v", resp)

		return resp
	},
		http.MethodGet,
	)
}

func RunSearch(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunSearch request: %#v", req)

		runs := []model.Run{}
		tx := db.Where("experiment_id IN ?", req.ExperimentIDs)

		// ViewType
		var lifecyleStages []model.LifecycleStage
		switch req.ViewType {
		case ViewTypeActiveOnly, "":
			lifecyleStages = []model.LifecycleStage{
				model.LifecycleStageActive,
			}
		case ViewTypeDeletedOnly:
			lifecyleStages = []model.LifecycleStage{
				model.LifecycleStageDeleted,
			}
		case ViewTypeAll:
			lifecyleStages = []model.LifecycleStage{
				model.LifecycleStageActive,
				model.LifecycleStageDeleted,
			}
		default:
			return NewError(ErrorCodeInvalidParameterValue, "Invalid run_view_type '%s'", req.ViewType)
		}
		tx.Where("lifecycle_stage IN ?", lifecyleStages)

		// MaxResults
		// TODO if compatible with mlflow client, consider using same logic as in ExperimentSearch
		limit := int(req.MaxResults)
		if limit == 0 {
			limit = 1000
		} else if limit > 1000000 {
			return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter 'max_results' supplied.")
		}
		tx.Limit(limit)

		// PageToken
		var offset int
		if req.PageToken != "" {
			var token PageToken
			if err := json.NewDecoder(
				base64.NewDecoder(
					base64.StdEncoding,
					strings.NewReader(req.PageToken),
				),
			).Decode(&token); err != nil {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid page_token '%s': %s", req.PageToken, err)

			}
			offset = int(token.Offset)
		}
		tx.Offset(offset)

		// Filter
		if req.Filter != "" {
			for n, f := range filterAnd.Split(req.Filter, -1) {
				components := filterCond.FindStringSubmatch(f)
				if len(components) != 5 {
					return NewError(ErrorCodeInvalidParameterValue, "Malformed filter '%s'", f)
				}

				entity := components[1]
				key := strings.Trim(components[2], "\"`")
				comparison := components[3]
				var value any = components[4]

				var kind interface{}
				switch entity {
				case "", "attribute", "attributes", "attr", "run":
					switch key {
					case "start_time", "end_time":
						switch comparison {
						case ">", ">=", "!=", "=", "<", "<=":
							v, err := strconv.Atoi(value.(string))
							if err != nil {
								return NewError(ErrorCodeInvalidParameterValue, "Invalid numeric value '%s'", value)
							}
							value = v
						default:
							return NewError(ErrorCodeInvalidParameterValue, "Invalid numeric attribute comparison operator '%s'", comparison)
						}
					case "run_name":
						key = "mlflow.runName"
						kind = &model.Tag{}
						fallthrough
					case "status", "user_id", "artifact_uri":
						switch strings.ToUpper(comparison) {
						case "!=", "=", "LIKE", "ILIKE":
							if strings.HasPrefix(value.(string), "(") {
								return NewError(ErrorCodeInvalidParameterValue, "Invalid string value '%s'", value)
							}
							value = strings.Trim(value.(string), `"'`)
						default:
							return NewError(ErrorCodeInvalidParameterValue, "Invalid string attribute comparison operator '%s'", comparison)
						}
					case "run_id":
						key = "run_uuid"
						switch strings.ToUpper(comparison) {
						case "!=", "=", "LIKE", "ILIKE":
							if strings.HasPrefix(value.(string), "(") {
								return NewError(ErrorCodeInvalidParameterValue, "Invalid string value '%s'", value)
							}
							value = strings.Trim(value.(string), `"'`)
						case "IN", "NOT IN":
							if !strings.HasPrefix(value.(string), "(") {
								return NewError(ErrorCodeInvalidParameterValue, "Invalid list definition '%s'", value)
							}
							var values []string
							for _, v := range filterInGroup.Split(value.(string)[1:len(value.(string))-1], -1) {
								values = append(values, strings.Trim(v, "'"))
							}
							value = values
						default:
							return NewError(ErrorCodeInvalidParameterValue, "Invalid string attribute comparison operator '%s'", comparison)
						}
					default:
						return NewError(ErrorCodeInvalidParameterValue, "Invalid attribute '%s'. Valid values are ['run_name', 'start_time', 'end_time', 'status', 'user_id', 'artifact_uri', 'run_id']", key)
					}
				case "metric", "metrics":
					switch comparison {
					case ">", ">=", "!=", "=", "<", "<=":
						v, err := strconv.ParseFloat(value.(string), 64)
						if err != nil {
							return NewError(ErrorCodeInvalidParameterValue, "Invalid numeric value '%s'", value)
						}
						value = v
					default:
						return NewError(ErrorCodeInvalidParameterValue, "Invalid metric comparison operator '%s'", comparison)
					}
					kind = &model.LatestMetric{}
				case "parameter", "parameters", "param", "params":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return NewError(ErrorCodeInvalidParameterValue, "Invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					default:
						return NewError(ErrorCodeInvalidParameterValue, "Invalid param comparison operator '%s'", comparison)
					}
					kind = &model.Param{}
				case "tag", "tags":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return NewError(ErrorCodeInvalidParameterValue, "Invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
					default:
						return NewError(ErrorCodeInvalidParameterValue, "Invalid tag comparison operator '%s'", comparison)
					}
					kind = &model.Tag{}
				default:
					return NewError(ErrorCodeInvalidParameterValue, "Invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", entity)
				}
				if kind == nil {
					if db.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
						key = fmt.Sprintf("LOWER(%s)", key)
						comparison = "LIKE"
						value = strings.ToLower(value.(string))
					}
					tx.Where(fmt.Sprintf("runs.%s %s ?", key, comparison), value)
				} else {
					table := fmt.Sprintf("filter_%d", n)
					where := fmt.Sprintf("value %s ?", comparison)
					if db.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
						where = "LOWER(value) LIKE ?"
						value = strings.ToLower(value.(string))
					}
					tx.Joins(
						fmt.Sprintf("JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
						db.Select("run_uuid", "value").Where("key = ?", key).Where(where, value).Model(kind),
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
				return NewError(ErrorCodeInvalidParameterValue, "Invalid order_by clause '%s'", o)
			}

			column := strings.Trim(components[2], "`\"")

			var kind interface{}
			switch components[1] {
			case "attribute":
				if column == "start_time" {
					startTimeOrder = true
				}
			case "metric":
				kind = &model.LatestMetric{}
			case "param":
				kind = &model.Param{}
			case "tag":
				kind = &model.Tag{}
			default:
				return NewError(ErrorCodeInvalidParameterValue, "Invalid entity type '%s'. Valid values are ['metric', 'parameter', 'tag', 'attribute']", components[1])
			}
			if kind != nil {
				table := fmt.Sprintf("order_%d", n)
				tx.Joins(
					fmt.Sprintf("LEFT OUTER JOIN (?) AS %s ON runs.run_uuid = %s.run_uuid", table, table),
					db.Select("run_uuid", "value").Where("key = ?", column).Model(kind),
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
		tx.Preload("LatestMetrics").
			Preload("Params").
			Preload("Tags").
			Find(&runs)
		if tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to search runs: %s", tx.Error)
		}

		resp := &RunSearchResponse{
			Runs: make([]Run, len(runs)),
		}
		for n, r := range runs {
			resp.Runs[n] = modelRunToAPI(r)
		}

		// NextPageToken
		if len(runs) == limit {
			var token strings.Builder
			b64 := base64.NewEncoder(base64.StdEncoding, &token)
			if err := json.NewEncoder(b64).Encode(PageToken{
				Offset: int32(offset + limit),
			}); err != nil {
				return NewError(ErrorCodeInternalError, "Unable to build next_page_token: %s", err)
			}
			b64.Close()
			resp.NextPageToken = token.String()
		}

		log.Debugf("RunSearch response: %#v", resp)

		return resp
	},
		http.MethodPost,
	))
}

func RunDelete(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunDelete request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		run := model.Run{
			ID: req.ID,
		}
		if tx := db.Select("lifecycle_stage").First(&run); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find run '%s': %s", run.ID, tx.Error)
		}

		if tx := db.Model(&run).Updates(model.Run{
			LifecycleStage: model.LifecycleStageDeleted,
			DeletedTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to update run '%s': %s", run.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunRestore(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunRestoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunRestore request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		run := model.Run{
			ID: req.ID,
		}
		if tx := db.Select("lifecycle_stage").First(&run); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find run '%s': %s", run.ID, tx.Error)
		}

		// Use UpdateColumns so we can reset DeletedTime to null
		if tx := db.Model(&run).UpdateColumns(map[string]interface{}{
			"LifecycleStage": model.LifecycleStageActive,
			"DeletedTime":    sql.NullInt64{},
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to update run '%s': %s", run.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunLogMetric(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunLogMetricRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err, ok := err.(*json.UnmarshalTypeError); ok {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
			}
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunLogMetric request: %#v", req)

		if req.ID == "" && req.UUID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		if req.Key == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'key'")
		}

		// if req.Value == "" {
		// 	return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'value'")
		// }

		if req.Timestamp == 0 {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'timestamp'")
		}

		id := req.ID
		if id == "" {
			id = req.UUID
		}

		if err := runLogMetrics(db, id, []Metric{req.Metric}); err != nil {
			return NewError(ErrorCodeInternalError, "Unable to log metric '%s' for run '%s': %s", req.Key, id, err)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunLogParam(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunLogParamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err, ok := err.(*json.UnmarshalTypeError); ok {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
			}
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunLogParam request: %#v", req)

		if req.ID == "" && req.UUID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		if req.Key == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'key'")
		}

		id := req.ID
		if id == "" {
			id = req.UUID
		}

		if err := runLogParams(db, id, []RunParam{req.RunParam}); err != nil {
			return NewError(ErrorCodeInternalError, "Unable to log param '%s' for run '%s': %s", req.Key, id, err)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunSetTag(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunSetTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err, ok := err.(*json.UnmarshalTypeError); ok {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
			}
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunSetTag request: %#v", req)

		if req.ID == "" && req.UUID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		if req.Key == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'key'")
		}

		id := req.ID
		if id == "" {
			id = req.UUID
		}

		if err := runSetTags(db, id, []RunTag{req.RunTag}); err != nil {
			return NewError(ErrorCodeInternalError, "Unable to set tag '%s' for run '%s': %s", req.Key, id, err)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunDeleteTag(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunDeleteTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunDeleteTag request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		if tx := db.Select("run_uuid").Where("lifecycle_stage = ?", model.LifecycleStageActive).First(&model.Run{ID: req.ID}); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find active run '%s': %s", req.ID, tx.Error)
		}

		if tx := db.First(&model.Tag{RunID: req.ID, Key: req.Key}); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find tag '%s' for run '%s': %s", req.Key, req.ID, tx.Error)
		}

		if tx := db.Delete(&model.Tag{
			RunID: req.ID,
			Key:   req.Key,
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to delete tag '%s' for run '%s': %s", req.Key, req.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func RunLogBatch(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req RunLogBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err, ok := err.(*json.UnmarshalTypeError); ok {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
			}
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("RunLogBatch request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		if err := runLogParams(db, req.ID, req.Params); err != nil {
			return err
		}

		if err := runLogMetrics(db, req.ID, req.Metrics); err != nil {
			return err
		}

		if err := runSetTags(db, req.ID, req.Tags); err != nil {
			return err
		}

		return nil
	},
		http.MethodPost,
	))
}

func runLogMetrics(db *gorm.DB, id string, metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	if tx := db.Select("run_uuid").Where("lifecycle_stage = ?", model.LifecycleStageActive).First(&model.Run{ID: id}); tx.Error != nil {
		return NewError(ErrorCodeResourceDoesNotExist, "Unable to find active run '%s': %s", id, tx.Error)
	}

	dbMetrics := make([]model.Metric, len(metrics))
	latestMetrics := make(map[string]model.LatestMetric)
	for n, metric := range metrics {
		m := model.Metric{
			RunID:     id,
			Key:       metric.Key,
			Timestamp: metric.Timestamp,
			Step:      metric.Step,
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
				return NewError(ErrorCodeInvalidParameterValue, "Invalid metric value '%s'", v)
			}
		} else {
			return NewError(ErrorCodeInvalidParameterValue, "Invalid metric value '%s'", v)
		}
		dbMetrics[n] = m

		lm, ok := latestMetrics[m.Key]
		if !ok ||
			m.Step > lm.Step ||
			(m.Step == lm.Step && m.Timestamp > lm.Timestamp) ||
			(m.Step == lm.Step && m.Timestamp == lm.Timestamp && m.Value > lm.Value) {
			latestMetrics[m.Key] = model.LatestMetric{
				RunID:     m.RunID,
				Key:       m.Key,
				Value:     m.Value,
				Timestamp: m.Timestamp,
				Step:      m.Step,
				IsNan:     m.IsNan,
			}
		}
	}

	if tx := db.Create(&dbMetrics); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to insert metrics for run '%s': %s", id, tx.Error)
	}

	// TODO update latest metrics in the background?

	keys := make([]string, len(latestMetrics))
	n := 0
	for k := range latestMetrics {
		keys[n] = k
		n += 1
	}

	var currentLatestMetrics []model.LatestMetric
	if tx := db.Where("run_uuid = ?", id).Where("key IN ?", keys).Find(&currentLatestMetrics); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to get latest metrics for run '%s': %s", id, tx.Error)
	}

	currentLatestMetricsMap := make(map[string]model.LatestMetric, len(currentLatestMetrics))
	for _, m := range currentLatestMetrics {
		currentLatestMetricsMap[m.Key] = m
	}

	updatedLatestMetrics := make([]model.LatestMetric, 0, len(keys))
	for k, m := range latestMetrics {
		lm, ok := currentLatestMetricsMap[k]
		if !ok ||
			m.Step > lm.Step ||
			(m.Step == lm.Step && m.Timestamp > lm.Timestamp) ||
			(m.Step == lm.Step && m.Timestamp == lm.Timestamp && m.Value > lm.Value) {
			updatedLatestMetrics = append(updatedLatestMetrics, m)
		}
	}

	if len(updatedLatestMetrics) > 0 {
		if tx := db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&updatedLatestMetrics); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to update latest metrics for run '%s': %s", id, tx.Error)
		}
	}

	return nil
}

func runLogParams(db *gorm.DB, id string, params []RunParam) error {
	if len(params) == 0 {
		return nil
	}

	if tx := db.Select("run_uuid").Where("lifecycle_stage = ?", model.LifecycleStageActive).First(&model.Run{ID: id}); tx.Error != nil {
		return NewError(ErrorCodeResourceDoesNotExist, "Unable to find active run '%s': %s", id, tx.Error)
	}

	dbParams := make([]model.Param, len(params))
	for n, p := range params {
		dbParams[n] = model.Param{
			Key:   p.Key,
			Value: p.Value,
			RunID: id,
		}
	}

	if tx := db.Create(&dbParams); tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to insert params for run '%s': %s", id, tx.Error)
	}

	return nil
}

func runSetTags(db *gorm.DB, id string, tags []RunTag) error {
	if len(tags) == 0 {
		return nil
	}

	run := model.Run{ID: id}
	if tx := db.Select("run_uuid", "name", "user_id").Where("lifecycle_stage = ?", model.LifecycleStageActive).First(&run); tx.Error != nil {
		return NewError(ErrorCodeResourceDoesNotExist, "Unable to find active run '%s': %s", id, tx.Error)
	}

	tx := db.Begin()
	dbTags := make([]model.Tag, len(tags))
	for n, t := range tags {
		dbTags[n] = model.Tag{
			Key:   t.Key,
			Value: t.Value,
			RunID: id,
		}
		switch t.Key {
		case "mlflow.runName":
			if run.Name != t.Value {
				tx.Model(&run).UpdateColumn("name", t.Value)
			}
		case "mlflow.user":
			if run.UserID != t.Value {
				tx.Model(&run).UpdateColumn("user_id", t.Value)
			}
		}
	}

	tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&dbTags)

	tx.Commit()
	if tx.Error != nil {
		return NewError(ErrorCodeInternalError, "Unable to insert tags for run '%s': %s", id, tx.Error)
	}

	return nil
}

func modelRunToAPI(r model.Run) Run {
	metrics := make([]Metric, len(r.LatestMetrics))
	for n, m := range r.LatestMetrics {

		metrics[n] = Metric{
			Key:       m.Key,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Step:      m.Step,
		}
		if m.IsNan {
			metrics[n].Value = "NaN"
		}
	}

	params := make([]RunParam, len(r.Params))
	for n, p := range r.Params {
		params[n] = RunParam{
			Key:   p.Key,
			Value: p.Value,
		}
	}

	tags := make([]RunTag, len(r.Tags))
	for n, t := range r.Tags {
		tags[n] = RunTag{
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

	return Run{
		RunInfo{
			ID:             r.ID,
			UUID:           r.ID,
			Name:           r.Name,
			ExperimentID:   fmt.Sprint(r.ExperimentID),
			UserID:         r.UserID,
			Status:         RunStatus(r.Status),
			StartTime:      r.StartTime.Int64,
			EndTime:        r.EndTime.Int64,
			ArtifactURI:    r.ArtifactURI,
			LifecycleStage: LifecycleStage(r.LifecycleStage),
		},
		RunData{
			Metrics: metrics,
			Params:  params,
			Tags:    tags,
		},
	}

}
