package mlflow

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrack/pkg/database"
)

var (
	experimentOrder = regexp.MustCompile(`^(?:attr(?:ibutes?)?\.)?(\w+)(?i:\s+(ASC|DESC))?$`)
)

func CreateExperiment(c *fiber.Ctx) error {
	var req request.CreateExperimentRequest
	if err := c.BodyParser(&req); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return api.NewCodeInvalidParameterValueError("Invalid value for parameter '%s' supplied. Hint: Value was of type '%s'. See the API docs for more information about request parameters.", err.Field, err.Value)
		}
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("CreateExperiment request: %#v", &req)

	if req.Name == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'name'")
	}

	if req.ArtifactLocation != "" {
		u, err := url.Parse(req.ArtifactLocation)
		if err != nil {
			return api.NewCodeInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
		}

		p, err := filepath.Abs(u.Path)
		if err != nil {
			return api.NewCodeInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
		}
		u.Path = p
		req.ArtifactLocation = u.String()
	}

	ts := time.Now().UTC().UnixMilli()
	exp := database.Experiment{
		Name:             req.Name,
		ArtifactLocation: req.ArtifactLocation,
		LifecycleStage:   database.LifecycleStageActive,
		CreationTime: sql.NullInt64{
			Int64: ts,
			Valid: true,
		},
		LastUpdateTime: sql.NullInt64{
			Int64: ts,
			Valid: true,
		},
		Tags: make([]database.ExperimentTag, len(req.Tags)),
	}

	for n, tag := range req.Tags {
		exp.Tags[n] = database.ExperimentTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}

	if tx := database.DB.Create(&exp); tx.Error != nil {
		if err, ok := tx.Error.(*pgconn.PgError); ok && err.Code == "23505" {
			return api.NewResourceAlreadyExistError("Experiment(name=%s) already exists", exp.Name)
		}
		if err, ok := tx.Error.(sqlite3.Error); ok && err.Code == 19 && err.ExtendedCode == 2067 {
			return api.NewResourceAlreadyExistError("Experiment(name=%s) already exists", exp.Name)
		}
		return api.NewInternalServerError("Error inserting experiment '%s': %s", exp.Name, tx.Error)
	}

	if exp.ArtifactLocation == "" {
		exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(viper.GetString("artifact-root"), "/"), *exp.ID)
		if tx := database.DB.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
			return api.NewInternalServerError("Error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
		}
	}

	resp := &response.CreateExperimentResponse{
		ID: fmt.Sprint(*exp.ID),
	}

	log.Debugf("CreateExperiment response: %#v", resp)

	return c.JSON(resp)
}

func UpdateExperiment(c *fiber.Ctx) error {
	var req request.UpdateExperimentRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("UpdateExperiment request: %#v", &req)
	if req.ID == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	if req.Name == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'new_name'")
	}

	ex, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID: &ex32,
	}

	if tx := database.DB.Select("ID").First(&exp); tx.Error != nil {
		return api.NewResourceNoExistsError("Unable to find experiment '%d': %s", *exp.ID, tx.Error)
	}

	if tx := database.DB.Model(&exp).Updates(&database.Experiment{
		Name: req.Name,
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
	}); tx.Error != nil {
		return api.NewInternalServerError("Unable to update experiment '%d': %s", *exp.ID, tx.Error)
	}

	return c.JSON(fiber.Map{})
}

func GetExperiment(c *fiber.Ctx) error {
	id := c.Query("experiment_id")

	log.Debugf("GetExperiment request: experiment_id='%s'", id)

	if id == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	ex, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", id, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID: &ex32,
	}

	if tx := database.DB.Preload("Tags").First(&exp); tx.Error != nil {
		return api.NewResourceNoExistsError("Unable to find experiment '%d': %s", ex, tx.Error)
	}

	resp := response.GetExperimentResponse{
		Experiment: response.ExperimentPartialResponse{
			ID:               fmt.Sprint(*exp.ID),
			Name:             exp.Name,
			ArtifactLocation: exp.ArtifactLocation,
			LifecycleStage:   string(exp.LifecycleStage),
			LastUpdateTime:   exp.LastUpdateTime.Int64,
			CreationTime:     exp.CreationTime.Int64,
			Tags:             make([]response.ExperimentTagPartialResponse, len(exp.Tags)),
		},
	}

	for n, t := range exp.Tags {
		resp.Experiment.Tags[n] = response.ExperimentTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
	}

	log.Debugf("GetExperiment response: %#v", resp)

	return c.JSON(resp)
}

func GetExperimentByName(c *fiber.Ctx) error {
	name := c.Query("experiment_name")

	log.Debugf("GetExperimentByName request: experiment_name='%s'", name)

	if name == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_name'")
	}

	exp := database.Experiment{
		Name: name,
	}
	if tx := database.DB.Preload("Tags").Where(&exp).First(&exp); tx.Error != nil {
		return api.NewResourceNoExistsError("Unable to find experiment '%s': %s", name, tx.Error)
	}

	resp := response.GetExperimentResponse{
		Experiment: response.ExperimentPartialResponse{
			ID:               fmt.Sprint(*exp.ID),
			Name:             exp.Name,
			ArtifactLocation: exp.ArtifactLocation,
			LifecycleStage:   string(exp.LifecycleStage),
			LastUpdateTime:   exp.LastUpdateTime.Int64,
			CreationTime:     exp.CreationTime.Int64,
			Tags:             make([]response.ExperimentTagPartialResponse, len(exp.Tags)),
		},
	}

	for n, t := range exp.Tags {
		resp.Experiment.Tags[n] = response.ExperimentTagPartialResponse{
			Key:   t.Key,
			Value: t.Value,
		}
	}

	log.Debugf("GetExperimentByName response: %#v", resp)

	return c.JSON(resp)
}

func DeleteExperiment(c *fiber.Ctx) error {
	var req request.DeleteExperimentRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("DeleteExperiment request: %#v", req)

	if req.ID == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	ex, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID: &ex32,
	}
	if tx := database.DB.Select("ID").First(&exp); tx.Error != nil {
		return api.NewResourceNoExistsError("Unable to find experiment '%d': %s", *exp.ID, tx.Error)
	}

	if tx := database.DB.Model(&exp).Updates(&database.Experiment{
		LifecycleStage: database.LifecycleStageDeleted,
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
	}); tx.Error != nil {
		return api.NewInternalServerError("Unable to delete experiment '%d': %s", *exp.ID, tx.Error)
	}

	return c.JSON(fiber.Map{})
}

func RestoreExperiment(c *fiber.Ctx) error {
	var req request.RestoreExperimentRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("RestoreExperiment request: %#v", req)

	if req.ID == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	ex, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID: &ex32,
	}
	if tx := database.DB.Select("ID").First(&exp); tx.Error != nil {
		return api.NewResourceNoExistsError("Unable to find experiment '%d': %s", *exp.ID, tx.Error)
	}

	if tx := database.DB.Model(&exp).Updates(&database.Experiment{
		LifecycleStage: database.LifecycleStageActive,
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
	}); tx.Error != nil {

		return api.NewInternalServerError("Unable to restore experiment '%d': %s", *exp.ID, tx.Error)
	}

	return c.JSON(fiber.Map{})
}

func SetExperimentTag(c *fiber.Ctx) error {
	var req request.SetExperimentTagRequest
	if err := c.BodyParser(&req); err != nil {
		return api.NewBadRequestError("Unable to decode request body: %s", err)
	}

	log.Debugf("SetExperimentTag request: %#v", req)

	if req.ID == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	if req.Key == "" {
		return api.NewCodeInvalidParameterValueError("Missing value for required parameter 'key'")
	}

	ex, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	ex32 := int32(ex)
	exp := database.Experiment{
		ID:             &ex32,
		LifecycleStage: database.LifecycleStageActive,
	}
	if tx := database.DB.Select("ID").Where(&exp).First(&exp); tx.Error != nil {
		return api.NewCodeInvalidParameterValueError("Unable to find experiment '%d': %s", *exp.ID, tx.Error)
	}

	if tx := database.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&database.ExperimentTag{
		ExperimentID: *exp.ID,
		Key:          req.Key,
		Value:        req.Value,
	}); tx.Error != nil {
		return api.NewInternalServerError("Unable to set tag for experiment '%d': %s", *exp.ID, tx.Error)
	}

	return c.JSON(fiber.Map{})
}

func SearchExperiments(c *fiber.Ctx) error {
	var req request.SearchExperimentsRequest
	switch c.Method() {
	case fiber.MethodPost:
		if err := c.BodyParser(&req); err != nil {
			return api.NewBadRequestError("Unable to decode request body: %s", err)
		}
	case fiber.MethodGet:
		if err := c.QueryParser(&req); err != nil {
			return api.NewBadRequestError(err.Error())
		}
	}

	log.Debugf("SearchExperiments request: %#v", req)

	exps := []database.Experiment{}

	// ViewType
	var lifecyleStages []database.LifecycleStage
	switch req.ViewType {
	case string(request.ViewTypeActiveOnly), "":
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageActive,
		}
	case string(request.ViewTypeDeletedOnly):
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageDeleted,
		}
	case string(request.ViewTypeAll):
		lifecyleStages = []database.LifecycleStage{
			database.LifecycleStageActive,
			database.LifecycleStageDeleted,
		}
	default:
		return api.NewCodeInvalidParameterValueError("Invalid view_type '%s'", req.ViewType)
	}
	tx := database.DB.Where("lifecycle_stage IN ?", lifecyleStages)

	// MaxResults
	limit := int(req.MaxResults)
	if limit == 0 {
		limit = 1000
	}
	tx.Limit(limit + 1)

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
			return api.NewCodeInvalidParameterValueError("Invalid page_token '%s': %s", req.PageToken, err)

		}
		offset = int(token.Offset)
	}
	tx.Offset(offset)

	// Filter
	if req.Filter != "" {
		for n, f := range filterAnd.Split(req.Filter, -1) {
			components := filterCond.FindStringSubmatch(f)
			if len(components) != 5 {
				return api.NewCodeInvalidParameterValueError("Malformed filter '%s'", f)
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
							return api.NewCodeInvalidParameterValueError("Invalid numeric value '%s'", value)
						}
						value = v
					default:
						return api.NewCodeInvalidParameterValueError("Invalid numeric attribute comparison operator '%s'", comparison)
					}
				case "name":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return api.NewCodeInvalidParameterValueError("Invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
						if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
							key = fmt.Sprintf("LOWER(%s)", key)
							comparison = "LIKE"
							value = strings.ToLower(value.(string))
						}
					default:
						return api.NewCodeInvalidParameterValueError("Invalid string attribute comparison operator '%s'", comparison)
					}
				default:
					return api.NewCodeInvalidParameterValueError("Invalid attribute '%s'. Valid values are ['name', 'creation_time', 'last_update_time']", key)
				}
				tx.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
			case "tag", "tags":
				switch strings.ToUpper(comparison) {
				case "!=", "=", "LIKE", "ILIKE":
					if strings.HasPrefix(value.(string), "(") {
						return api.NewCodeInvalidParameterValueError("Invalid string value '%s'", value)
					}
					value = strings.Trim(value.(string), `"'`)
				default:
					return api.NewCodeInvalidParameterValueError("Invalid tag comparison operator '%s'", comparison)
				}
				table := fmt.Sprintf("filter_%d", n)
				where := fmt.Sprintf("value %s ?", comparison)
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
					where = "LOWER(value) LIKE ?"
					value = strings.ToLower(value.(string))
				}
				tx.Joins(
					fmt.Sprintf("JOIN (?) AS %s ON experiments.experiment_id = %s.experiment_id", table, table),
					database.DB.Select("experiment_id", "value").Where("key = ?", key).Where(where, value).Model(&database.ExperimentTag{}),
				)
			default:
				return api.NewCodeInvalidParameterValueError("Invalid entity type '%s'. Valid values are ['tag', 'attribute']", entity)
			}
		}
	}

	// OrderBy
	expOrder := false
	for _, o := range req.OrderBy {
		components := experimentOrder.FindStringSubmatch(o)
		if len(components) == 0 {
			return api.NewCodeInvalidParameterValueError("Invalid order_by clause '%s'", o)
		}

		column := components[1]
		switch column {
		case "experiment_id":
			expOrder = true
			fallthrough
		case "name", "creation_time", "last_update_time":
		default:
			return api.NewCodeInvalidParameterValueError("Invalid attribute '%s'. Valid values are ['name', 'experiment_id', 'creation_time', 'last_update_time']", column)
		}
		tx.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
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
		return api.NewInternalServerError("Unable to search runs: %s", tx.Error)
	}

	resp := &response.SearchExperimentsResponse{}

	// NextPageToken
	if len(exps) > limit {
		exps = exps[:limit]
		var token strings.Builder
		b64 := base64.NewEncoder(base64.StdEncoding, &token)
		if err := json.NewEncoder(b64).Encode(request.PageToken{
			Offset: int32(offset + limit),
		}); err != nil {
			return api.NewInternalServerError("Unable to build next_page_token: %s", err)
		}
		b64.Close()
		resp.NextPageToken = token.String()
	}

	resp.Experiments = make([]response.ExperimentPartialResponse, len(exps))
	for n, r := range exps {
		e := response.ExperimentPartialResponse{
			ID:               fmt.Sprint(*r.ID),
			Name:             r.Name,
			ArtifactLocation: r.ArtifactLocation,
			LifecycleStage:   string(r.LifecycleStage),
			LastUpdateTime:   r.LastUpdateTime.Int64,
			CreationTime:     r.CreationTime.Int64,
			Tags:             make([]response.ExperimentTagPartialResponse, len(r.Tags)),
		}

		for n, t := range r.Tags {
			e.Tags[n] = response.ExperimentTagPartialResponse{
				Key:   t.Key,
				Value: t.Value,
			}
		}

		resp.Experiments[n] = e
	}

	log.Debugf("SearchExperiments response: %#v", resp)

	return c.JSON(resp)
}
