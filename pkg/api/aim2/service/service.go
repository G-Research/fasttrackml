package service

import (
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
)

// Service provides service layer to work with `run` business logic.
type Service struct {
	runRepository        repositories.RunRepositoryProvider
	paramRepository      repositories.ParamRepositoryProvider
	metricRepository     repositories.MetricRepositoryProvider
	experimentRepository repositories.ExperimentRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	runRepository repositories.RunRepositoryProvider,
	paramRepository repositories.ParamRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
	experimentRepository repositories.ExperimentRepositoryProvider,
) *Service {
	return &Service{
		runRepository:        runRepository,
		paramRepository:      paramRepository,
		metricRepository:     metricRepository,
		experimentRepository: experimentRepository,
	}
}
