package experiment

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

var (
	experimentOrder = regexp.MustCompile(`^(?:attr(?:ibutes?)?\.)?(\w+)(?i:\s+(ASC|DESC))?$`)
	filterAnd       = regexp.MustCompile(`(?i)\s+AND\s+`)
	filterCond      = regexp.MustCompile(`^(?:(\w+)\.)?("[^"]+"|` + "`[^`]+`" + `|[\w\.]+)\s+(<|<=|>|>=|=|!=|(?i:I?LIKE)|(?i:(?:NOT )?IN))\s+(\((?:'[^']+'(?:,\s*)?)+\)|"[^"]+"|'[^']+'|[\w\.]+)$`)
)

// Service provides service layer to work with `metric` business logic.
type Service struct {
	config               *config.ServiceConfig
	tagRepository        repositories.TagRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	config *config.ServiceConfig,
	tagRepository repositories.TagRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		config:               config,
		tagRepository:        tagRepository,
		experimentRepository: experimentRepository,
	}
}

// CreateExperiment creates new Experiment entity.
func (s Service) CreateExperiment(
	ctx context.Context, ns *models.Namespace, req *request.CreateExperimentRequest,
) (*models.Experiment, error) {
	if err := ValidateCreateExperimentRequest(req); err != nil {
		return nil, err
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndName(ctx, ns.ID, req.Name)
	if err != nil {
		return nil, api.NewInternalError("error getting experiment with name: '%s', error: %s", req.Name, err)
	}
	if experiment != nil {
		return nil, api.NewResourceAlreadyExistsError("experiment(name=%s) already exists", req.Name)
	}

	experiment, err = convertors.ConvertCreateExperimentToDBModel(req)
	if err != nil {
		return nil, api.NewInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
	}
	experiment.NamespaceID = ns.ID

	if err := s.experimentRepository.Create(ctx, experiment); err != nil {
		return nil, api.NewInternalError("error inserting experiment '%s': %s", req.Name, err)
	}

	if experiment.ArtifactLocation == "" {
		path, err := url.JoinPath(s.config.ArtifactRoot, fmt.Sprintf("%d", *experiment.ID))
		if err != nil {
			return nil, api.NewInternalError(
				"error creating artifact_location for experiment'%s': %s", experiment.Name, err,
			)
		}
		experiment.ArtifactLocation = path
		if err := s.experimentRepository.Update(ctx, experiment); err != nil {
			return nil, api.NewInternalError(
				"error updating artifact_location for experiment '%s': %s", experiment.Name, err,
			)
		}
	}

	return experiment, nil
}

// UpdateExperiment updates existing Experiment entity.
func (s Service) UpdateExperiment(
	ctx context.Context, ns *models.Namespace, req *request.UpdateExperimentRequest,
) error {
	if err := ValidateUpdateExperimentRequest(req); err != nil {
		return err
	}

	parsedID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("unable to parse experiment id '%s': %s", req.ID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(parsedID))
	if err != nil {
		return api.NewResourceDoesNotExistError("unable to find experiment '%d': %s", parsedID, err)
	}

	experiment = convertors.ConvertUpdateExperimentToDBModel(experiment, req)
	if err := s.experimentRepository.Update(ctx, experiment); err != nil {
		return api.NewInternalError("unable to update experiment '%d': %s", *experiment.ID, err)
	}

	return nil
}

// GetExperiment returns existing Experiment entity by ID.
func (s Service) GetExperiment(
	ctx context.Context, ns *models.Namespace, req *request.GetExperimentRequest,
) (*models.Experiment, error) {
	if err := ValidateGetExperimentByIDRequest(req); err != nil {
		return nil, err
	}

	// TODO:DSuhinin not sure about this conversion. Maybe we can just use string everywhere?
	parsedID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return nil, api.NewBadRequestError(`unable to parse experiment id '%s': %s`, req.ID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(parsedID))
	if err != nil {
		return nil, api.NewResourceDoesNotExistError(`unable to find experiment '%d': %s`, parsedID, err)
	}

	return experiment, nil
}

// GetExperimentByName returns existing Experiment entity by Name.
func (s Service) GetExperimentByName(
	ctx context.Context, ns *models.Namespace, req *request.GetExperimentRequest,
) (*models.Experiment, error) {
	if err := ValidateGetExperimentByNameRequest(req); err != nil {
		return nil, err
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndName(ctx, ns.ID, req.Name)
	if err != nil {
		return nil, api.NewInternalError("unable to get experiment by name '%s': %v", req.Name, err)
	}
	if experiment == nil {
		return nil, api.NewResourceDoesNotExistError(`unable to find experiment '%s'`, req.Name)
	}

	return experiment, nil
}

// DeleteExperiment deletes existing Experiment entity.
func (s Service) DeleteExperiment(
	ctx context.Context, ns *models.Namespace, req *request.DeleteExperimentRequest,
) error {
	if err := ValidateDeleteExperimentRequest(req); err != nil {
		return err
	}

	parsedID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("unable to parse experiment id '%s': %s", req.ID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(parsedID))
	if err != nil {
		return api.NewResourceDoesNotExistError(`unable to find experiment '%d': %s`, parsedID, err)
	}

	experiment.LifecycleStage = models.LifecycleStageDeleted
	experiment.LastUpdateTime = sql.NullInt64{
		Int64: time.Now().UTC().UnixMilli(),
		Valid: true,
	}

	if err := s.experimentRepository.Update(ctx, experiment); err != nil {
		return api.NewInternalError("unable to delete experiment '%d': %s", *experiment.ID, err)
	}

	return nil
}

// RestoreExperiment restores deleted Experiment entity.
func (s Service) RestoreExperiment(
	ctx context.Context, ns *models.Namespace, req *request.RestoreExperimentRequest,
) error {
	if err := ValidateRestoreExperimentRequest(req); err != nil {
		return err
	}

	parsedID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(parsedID))
	if err != nil {
		return api.NewResourceDoesNotExistError(`unable to find experiment '%d': %s`, parsedID, err)
	}

	experiment.LifecycleStage = models.LifecycleStageActive
	experiment.LastUpdateTime = sql.NullInt64{
		Int64: time.Now().UTC().UnixMilli(),
		Valid: true,
	}

	if err := s.experimentRepository.Update(ctx, experiment); err != nil {
		return api.NewInternalError("Unable to restore experiment '%d': %s", *experiment.ID, err)
	}

	return nil
}

func (s Service) SetExperimentTag(
	ctx context.Context, ns *models.Namespace, req *request.SetExperimentTagRequest,
) error {
	if err := ValidateSetExperimentTagRequest(req); err != nil {
		return err
	}

	parsedID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", req.ID, err)
	}

	experiment, err := s.experimentRepository.GetByNamespaceIDAndExperimentID(ctx, ns.ID, int32(parsedID))
	if err != nil {
		return api.NewResourceDoesNotExistError(`unable to find experiment '%d': %s`, parsedID, err)
	}

	experimentTag := convertors.ConvertSetExperimentTagRequestToDBModel(*experiment.ID, req)
	if err := s.tagRepository.CreateExperimentTag(ctx, experimentTag); err != nil {
		return api.NewInternalError("Unable to set tag for experiment '%d': %s", *experiment.ID, err)
	}

	return nil
}

func (s Service) SearchExperiments(
	ctx context.Context, ns *models.Namespace, req *request.SearchExperimentsRequest,
) ([]models.Experiment, int, int, error) {
	if err := ValidateSearchExperimentsRequest(req); err != nil {
		return nil, 0, 0, err
	}

	query := database.DB.Where(
		"experiments.namespace_id = ?", ns.ID,
	)

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
	query.Where("lifecycle_stage IN ?", lifecyleStages)

	// MaxResults
	limit := int(req.MaxResults)
	if limit == 0 {
		limit = 1000
	}
	query.Limit(limit + 1)

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
	query.Offset(offset)

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

			switch entity {
			case "", "attribute", "attributes", "attr":
				switch key {
				case "creation_time", "last_update_time":
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
				case "name":
					switch strings.ToUpper(comparison) {
					case "!=", "=", "LIKE", "ILIKE":
						if strings.HasPrefix(value.(string), "(") {
							return nil, 0, 0, api.NewInvalidParameterValueError("invalid string value '%s'", value)
						}
						value = strings.Trim(value.(string), `"'`)
						if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
							key = fmt.Sprintf("LOWER(%s)", key)
							comparison = "LIKE"
							value = strings.ToLower(value.(string))
						}
					default:
						return nil, 0, 0, api.NewInvalidParameterValueError("invalid string attribute comparison operator '%s'", comparison)
					}
				default:
					return nil, 0, 0, api.NewInvalidParameterValueError("invalid attribute '%s'. Valid values are ['name', 'creation_time', 'last_update_time']", key)
				}
				query.Where(fmt.Sprintf("%s %s ?", key, comparison), value)
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
				table := fmt.Sprintf("filter_%d", n)
				where := fmt.Sprintf("value %s ?", comparison)
				if database.DB.Dialector.Name() == "sqlite" && strings.ToUpper(comparison) == "ILIKE" {
					where = "LOWER(value) LIKE ?"
					value = strings.ToLower(value.(string))
				}
				query.Joins(
					fmt.Sprintf("JOIN (?) AS %s ON experiments.experiment_id = %s.experiment_id", table, table),
					database.DB.Select("experiment_id", "value").Where("key = ?", key).Where(where, value).Model(&database.ExperimentTag{}),
				)
			default:
				return nil, 0, 0, api.NewInvalidParameterValueError("invalid entity type '%s'. Valid values are ['tag', 'attribute']", entity)
			}
		}
	}

	// OrderBy
	expOrder := false
	for _, o := range req.OrderBy {
		components := experimentOrder.FindStringSubmatch(o)
		if len(components) == 0 {
			return nil, 0, 0, api.NewInvalidParameterValueError("invalid order_by clause '%s'", o)
		}

		column := components[1]
		switch column {
		case "experiment_id":
			expOrder = true
			fallthrough
		case "name", "creation_time", "last_update_time":
		default:
			return nil, 0, 0, api.NewInvalidParameterValueError("invalid attribute '%s'. Valid values are ['name', 'experiment_id', 'creation_time', 'last_update_time']", column)
		}
		query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   len(components) == 3 && strings.ToUpper(components[2]) == "DESC",
		})

	}
	if len(req.OrderBy) == 0 {
		query.Order("experiments.creation_time DESC")
	}
	if !expOrder {
		query.Order("experiments.experiment_id ASC")
	}

	// Actual query
	var exps []models.Experiment
	if err := query.Preload("Tags").Find(&exps).Error; err != nil {
		return nil, 0, 0, api.NewInternalError("unable to search runs: %s", err)
	}

	return exps, limit, offset, nil
}
