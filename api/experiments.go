package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fasttrack/model"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	experimentFilterOrder *regexp.Regexp = regexp.MustCompile(`^(?:attr(?:ibutes?)?\.)?(\w+)(?i:\s+(ASC|DESC))?$`)
)

func ExperimentCreate(db *gorm.DB, artifactRoot string) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err, ok := err.(*json.UnmarshalTypeError); ok {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
			}
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("ExperimentCreate request: %#v", &req)

		if req.Name == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'name'")
		}

		ts := time.Now().UTC().UnixMilli()
		exp := model.Experiment{
			Name:             req.Name,
			ArtifactLocation: req.ArtifactLocation,
			LifecycleStage:   model.LifecycleStageActive,
			CreationTime: sql.NullInt64{
				Int64: ts,
				Valid: true,
			},
			LastUpdateTime: sql.NullInt64{
				Int64: ts,
				Valid: true,
			},
			Tags: make([]model.ExperimentTag, len(req.Tags)),
		}

		for n, tag := range req.Tags {
			exp.Tags[n] = model.ExperimentTag{
				Key:   tag.Key,
				Value: tag.Value,
			}
		}

		// TODO do it in one session?
		if tx := db.Create(&exp); tx.Error != nil {
			if err, ok := tx.Error.(*pgconn.PgError); ok && err.Code == "23505" {
				return NewError(ErrorCodeResourceAlreadyExists, "An experiment already exists with the name '%s'", exp.Name)
			}
			return NewError(ErrorCodeInternalError, "Error inserting experiment '%s': %s", exp.Name, tx.Error)
		}

		if exp.ArtifactLocation == "" {
			exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(artifactRoot, "/"), *exp.ID)
		}
		if tx := db.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
		}

		resp := &ExperimentCreateResponse{
			ID: fmt.Sprint(*exp.ID),
		}

		log.Debugf("ExperimentCreate response: %#v", resp)

		return resp
	},
		http.MethodPost,
	))
}

func ExperimentUpdate(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("ExperimentUpdate request: %#v", &req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_id'")
		}

		if req.Name == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'new_name'")
		}

		ex, err := strconv.ParseInt(req.ID, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", req.ID, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID: &ex32,
		}

		if tx := db.Select("ID").First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", *exp.ID, tx.Error)
		}

		if tx := db.Model(&exp).Updates(&model.Experiment{
			Name: req.Name,
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to update experiment '%d': %s", *exp.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func ExperimentGet(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		id := r.URL.Query().Get("experiment_id")

		log.Debugf("ExperimentGet request: experiment_id='%s'", id)

		if id == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_id'")
		}

		ex, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", id, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID: &ex32,
		}

		if tx := db.Preload("Tags").First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", ex, tx.Error)
		}

		resp := ExperimentGetResponse{
			Experiment: Experiment{
				ID:               fmt.Sprint(*exp.ID),
				Name:             exp.Name,
				ArtifactLocation: exp.ArtifactLocation,
				LifecycleStage:   LifecycleStage(exp.LifecycleStage),
				LastUpdateTime:   exp.LastUpdateTime.Int64,
				CreationTime:     exp.CreationTime.Int64,
				Tags:             make([]ExperimentTag, len(exp.Tags)),
			},
		}

		for n, t := range exp.Tags {
			resp.Experiment.Tags[n] = ExperimentTag{
				Key:   t.Key,
				Value: t.Value,
			}
		}

		log.Debugf("ExperimentGet response: %#v", resp)

		return resp
	},
		http.MethodGet,
	)
}

func ExperimentGetByName(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		name := r.URL.Query().Get("experiment_name")

		log.Debugf("ExperimentGetByName request: experiment_name='%s'", name)

		if name == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_name'")
		}

		exp := model.Experiment{
			Name: name,
		}
		if tx := db.Preload("Tags").Where(&exp).First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%s': %s", name, tx.Error)
		}

		resp := ExperimentGetResponse{
			Experiment: Experiment{
				ID:               fmt.Sprint(*exp.ID),
				Name:             exp.Name,
				ArtifactLocation: exp.ArtifactLocation,
				LifecycleStage:   LifecycleStage(exp.LifecycleStage),
				LastUpdateTime:   exp.LastUpdateTime.Int64,
				CreationTime:     exp.CreationTime.Int64,
				Tags:             make([]ExperimentTag, len(exp.Tags)),
			},
		}

		for n, t := range exp.Tags {
			resp.Experiment.Tags[n] = ExperimentTag{
				Key:   t.Key,
				Value: t.Value,
			}
		}

		log.Debugf("ExperimentGetByName response: %#v", resp)

		return resp
	},
		http.MethodGet,
	)
}

func ExperimentDelete(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("ExperimentDelete request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_id'")
		}

		ex, err := strconv.ParseInt(req.ID, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", req.ID, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID: &ex32,
		}
		if tx := db.Select("ID").First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", *exp.ID, tx.Error)
		}

		if tx := db.Model(&exp).Updates(&model.Experiment{
			LifecycleStage: model.LifecycleStageDeleted,
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to delete experiment '%d': %s", *exp.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func ExperimentRestore(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentRestoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("ExperimentRestore request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_id'")
		}

		ex, err := strconv.ParseInt(req.ID, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", req.ID, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID: &ex32,
		}
		if tx := db.Select("ID").First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", *exp.ID, tx.Error)
		}

		if tx := db.Model(&exp).Updates(&model.Experiment{
			LifecycleStage: model.LifecycleStageActive,
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
		}); tx.Error != nil {

			return NewError(ErrorCodeInternalError, "Unable to restore experiment '%d': %s", *exp.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func ExperimentSetTag(db *gorm.DB) HandlerFunc {
	return EnsureJson(EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentSetTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
		}

		log.Debugf("ExperimentSetTag request: %#v", req)

		if req.ID == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'experiment_id'")
		}

		if req.Key == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'key'")
		}

		ex, err := strconv.ParseInt(req.ID, 10, 32)
		if err != nil {
			return NewError(ErrorCodeBadRequest, "Unable to parse experiment id '%s': %s", req.ID, err)
		}

		ex32 := int32(ex)
		exp := model.Experiment{
			ID:             &ex32,
			LifecycleStage: model.LifecycleStageActive,
		}
		if tx := db.Select("ID").Where(&exp).First(&exp); tx.Error != nil {
			return NewError(ErrorCodeResourceDoesNotExist, "Unable to find experiment '%d': %s", *exp.ID, tx.Error)
		}

		if tx := db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&model.ExperimentTag{
			ExperimentID: *exp.ID,
			Key:          req.Key,
			Value:        req.Value,
		}); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to set tag for experiment '%d': %s", *exp.ID, tx.Error)
		}

		return nil
	},
		http.MethodPost,
	))
}

func ExperimentSearch(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		var req ExperimentSearchRequest
		if r.Method == http.MethodPost {
			if r.Header.Get("Content-Type") != "application/json" {
				return NewError(ErrorCodeBadRequest, "Invalid Content-Type '%s'", r.Header.Get("Content-Type"))
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return NewError(ErrorCodeBadRequest, "Unable to decode request body: %s", err)
			}
		} else {
			q := r.URL.Query()
			var err error
			if m := q.Get("max_results"); m != "" {
				req.MaxResults, err = strconv.ParseInt(m, 10, 64)
				if err != nil {
					return NewError(ErrorCodeInvalidParameterValue, "Unable to parse max_results '%s': %s", m, err)
				}
			}
			req.PageToken = q.Get("page_token")
			req.Filter = q.Get("filter")
			req.OrderBy = q["order_by"]
			req.ViewType = ViewType(q.Get("view_type"))
		}

		log.Debugf("ExperimentSearch request: %#v", req)

		exps := []model.Experiment{}

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
			return NewError(ErrorCodeInvalidParameterValue, "Invalid view_type '%s'", req.ViewType)
		}
		tx := db.Where("lifecycle_stage IN ?", lifecyleStages)

		// MaxResults
		limit := int(req.MaxResults)
		if limit == 0 {
			limit = 1000
		}
		tx.Limit(limit + 1)

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

				switch entity {
				case "", "attribute", "attributes", "attr":
					switch key {
					case "creation_time", "last_update_time":
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
					case "name":
						switch strings.ToUpper(comparison) {
						case "!=", "=", "LIKE", "ILIKE":
							if strings.HasPrefix(value.(string), "(") {
								return NewError(ErrorCodeInvalidParameterValue, "Invalid string value '%s'", value)
							}
							value = strings.Trim(value.(string), `"'`)
							if db.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
								key = fmt.Sprintf("LOWER(%s)", key)
								comparison = "LIKE"
								value = strings.ToLower(value.(string))
							}
						default:
							return NewError(ErrorCodeInvalidParameterValue, "Invalid string attribute comparison operator '%s'", comparison)
						}
					default:
						return NewError(ErrorCodeInvalidParameterValue, "Invalid attribute '%s'. Valid values are ['name', 'creation_time', 'last_update_time']", key)
					}
					tx.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
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
					table := fmt.Sprintf("filter_%d", n)
					where := fmt.Sprintf("value %s ?", comparison)
					if db.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
						where = "LOWER(value) LIKE ?"
						value = strings.ToLower(value.(string))
					}
					tx.Joins(
						fmt.Sprintf("JOIN (?) AS %s ON experiments.experiment_id = %s.experiment_id", table, table),
						db.Select("experiment_id", "value").Where("key = ?", key).Where(where, value).Model(&model.ExperimentTag{}),
					)
				default:
					return NewError(ErrorCodeInvalidParameterValue, "Invalid entity type '%s'. Valid values are ['tag', 'attribute']", entity)
				}
			}
		}

		// OrderBy
		expOrder := false
		for _, o := range req.OrderBy {
			components := experimentFilterOrder.FindStringSubmatch(o)
			if len(components) == 0 {
				return NewError(ErrorCodeInvalidParameterValue, "Invalid order_by clause '%s'", o)
			}

			column := components[1]
			switch column {
			case "experiment_id":
				expOrder = true
				fallthrough
			case "name":
				if db.Dialector.Name() == "postgres" {
					column += ` COLLATE "C"`
				}
			case "creation_time", "last_update_time":
			default:
				return NewError(ErrorCodeInvalidParameterValue, "Invalid attribute '%s'. Valid values are ['name', 'experiment_id', 'creation_time', 'last_update_time']", column)
			}
			tx.Order(clause.OrderByColumn{
				Column: clause.Column{Name: column, Raw: true},
				Desc:   len(components) == 3 && strings.ToUpper(components[2]) == "DESC",
			})

		}
		if len(req.OrderBy) == 0 {
			tx.Order("experiments.creation_time DESC")
		}
		if !expOrder {
			tx.Order("experiments.experiment_id ASC")
		}

		// Actual query
		tx.Preload("Tags").Find(&exps)
		if tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to search runs: %s", tx.Error)
		}

		resp := &ExperimentSearchResponse{}

		// NextPageToken
		if len(exps) > limit {
			exps = exps[:limit]
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

		resp.Experiments = make([]Experiment, len(exps))
		for n, r := range exps {
			e := Experiment{
				ID:               fmt.Sprint(*r.ID),
				Name:             r.Name,
				ArtifactLocation: r.ArtifactLocation,
				LifecycleStage:   LifecycleStage(r.LifecycleStage),
				LastUpdateTime:   r.LastUpdateTime.Int64,
				CreationTime:     r.CreationTime.Int64,
				Tags:             make([]ExperimentTag, len(r.Tags)),
			}

			for n, t := range r.Tags {
				e.Tags[n] = ExperimentTag{
					Key:   t.Key,
					Value: t.Value,
				}
			}

			resp.Experiments[n] = e
		}

		log.Debugf("ExperimentSearch response: %#v", resp)

		return resp
	},
		http.MethodGet,
		http.MethodPost,
	)
}
