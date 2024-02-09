package namespace

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// Service provides service layer to work with `namespace` business logic.
type Service struct {
	config               *config.ServiceConfig
	namespaceRepository  repositories.NamespaceRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	config *config.ServiceConfig,
	namespaceRepository repositories.NamespaceRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		config:               config,
		namespaceRepository:  namespaceRepository,
		experimentRepository: experimentRepository,
	}
}

// ListNamespaces returns all namespaces.
func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, error) {
	namespaces, err := s.namespaceRepository.List(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error listing namespaces")
	}
	return namespaces, nil
}

// GetNamespace returns one namespace by ID.
func (s Service) GetNamespace(ctx context.Context, id uint) (*models.Namespace, error) {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return nil, eris.Wrap(err, "error getting namespace by id")
	}
	return namespace, nil
}

// CreateNamespace creates a new namespace and default experiment.
func (s Service) CreateNamespace(ctx context.Context, code, description string) (*models.Namespace, error) {
	if err := ValidateNamespace(code); err != nil {
		return nil, eris.Wrap(err, "error validating namespace")
	}

	namespace := &models.Namespace{
		Code:                code,
		Description:         description,
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	}
	if err := s.namespaceRepository.Create(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error creating namespace")
	}

	timestamp := time.Now().UTC().UnixMilli()
	experiment := models.Experiment{
		Name:           models.DefaultExperimentName,
		NamespaceID:    namespace.ID,
		CreationTime:   sql.NullInt64{Int64: timestamp, Valid: true},
		LifecycleStage: models.LifecycleStageActive,
		LastUpdateTime: sql.NullInt64{Int64: timestamp, Valid: true},
	}

	if err := s.experimentRepository.Create(ctx, &experiment); err != nil {
		return nil, eris.Wrap(err, "error creating experiment")
	}

	// update Namespace with correct DefaultExperimentID now that it is known
	namespace.DefaultExperimentID = experiment.ID
	if err := s.namespaceRepository.Update(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error setting namespace default experiment id during create")
	}

	// setup ArtifactLocation for default experiment.
	path, err := url.JoinPath(s.config.DefaultArtifactRoot, fmt.Sprintf("%d", *experiment.ID))
	if err != nil {
		return nil, api.NewInternalError(
			"error creating artifact_location for experiment'%s': %s", experiment.Name, err,
		)
	}
	experiment.ArtifactLocation = path
	if err := s.experimentRepository.Update(ctx, &experiment); err != nil {
		return nil, api.NewInternalError(
			"error updating artifact_location for experiment '%s': %s", experiment.Name, err,
		)
	}

	return namespace, nil
}

// UpdateNamespace updates the code and description fields.
func (s Service) UpdateNamespace(ctx context.Context, id uint, code, description string) (*models.Namespace, error) {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return nil, eris.Wrapf(err, "error finding namespace by id: %d", id)
	}
	if namespace == nil {
		return nil, eris.Errorf("namespace not found by id: %d", id)
	}
	if err := ValidateNamespace(code); err != nil {
		return nil, eris.Wrap(err, "error validating namespace code")
	}
	namespace.Code = code
	namespace.Description = description

	if err := s.namespaceRepository.Update(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error updating namespace")
	}
	return namespace, nil
}

// DeleteNamespace deletes the namespace.
func (s Service) DeleteNamespace(ctx context.Context, id uint) error {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "error finding namespace by id: %d", id)
	}
	if namespace == nil {
		return eris.Errorf("namespace not found by id: %d", id)
	}
	if namespace.IsDefault() {
		return eris.Errorf("unable to delete default namespace")
	}
	if err := s.namespaceRepository.Delete(ctx, namespace); err != nil {
		return eris.Wrap(err, "error deleting namespace")
	}
	return nil
}
