package run

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/common"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/dao/types"
)

// allowed batch actions.
const (
	BatchActionDelete  = "delete"
	BatchActionArchive = "archive"
	BatchActionRestore = "restore"
)

// Service provides service layer to work with `run` business logic.
type Service struct {
	runRepository    repositories.RunRepositoryProvider
	metricRepository repositories.MetricRepositoryProvider
	tagRepository    repositories.TagRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	runRepository repositories.RunRepositoryProvider,
	metricRepository repositories.MetricRepositoryProvider,
	tagRepository repositories.TagRepositoryProvider,
) *Service {
	return &Service{
		runRepository:    runRepository,
		metricRepository: metricRepository,
		tagRepository:    tagRepository,
	}
}

// GetRunInfo returns run info.
func (s Service) GetRunInfo(
	ctx context.Context, namespaceID uint, req *request.GetRunInfoRequest,
) (*models.Run, error) {
	req = NormaliseGetRunInfoRequest(req)
	if err := ValidateGetRunInfoRequest(req); err != nil {
		return nil, err
	}

	runInfo, err := s.runRepository.GetRunInfo(ctx, namespaceID, req)
	if err != nil {
		return nil, api.NewInternalError("unable to find run by id %s: %s", req.ID, err)
	}
	if runInfo == nil {
		return nil, api.NewResourceDoesNotExistError("run '%s' not found", req.ID)
	}

	return runInfo, nil
}

// GetRunMetrics returns run metrics.
func (s Service) GetRunMetrics(
	ctx context.Context, namespaceID uint, runID string, req *request.GetRunMetricsRequest,
) ([]models.Metric, models.MetricKeysMap, error) {
	run, err := s.runRepository.GetRunByNamespaceIDAndRunID(ctx, namespaceID, runID)
	if err != nil {
		return nil, nil, api.NewInternalError("error getting run by id %s: %s", runID, err)
	}

	if run == nil {
		return nil, nil, api.NewResourceDoesNotExistError("run '%s' not found", runID)
	}

	metricKeysMap, err := ConvertRunMetricsRequestToMap(req)
	if err != nil {
		return nil, nil, api.NewBadRequestError("unable to convert request: %s", err)
	}
	metrics, err := s.runRepository.GetRunMetrics(ctx, runID, metricKeysMap)
	if err != nil {
		return nil, nil, api.NewInternalError("error getting run metrics by id %s: %s", runID, err)
	}

	return metrics, metricKeysMap, nil
}

// GetRunsActive returns the active runs.
func (s Service) GetRunsActive(
	ctx context.Context, namespaceID uint, req *request.GetRunsActiveRequest,
) ([]models.Run, error) {
	runs, err := s.runRepository.GetByNamespaceIDAndStatus(ctx, namespaceID, models.StatusRunning)
	if err != nil {
		return nil, api.NewInternalError("error ative runs: %s", err)
	}
	return runs, nil
}

// SearchRuns returns the list of runs by provided search criteria.
func (s Service) SearchRuns(
	ctx context.Context, namespaceID uint, tzOffset int, req request.SearchRunsRequest,
) ([]models.Run, int64, error) {
	runs, total, err := s.runRepository.SearchRuns(ctx, namespaceID, tzOffset, req)
	if err != nil {
		return nil, 0, api.NewInternalError("error searching runs: %s", err)
	}
	return runs, total, nil
}

// SearchMetrics returns the list of metrics by provided search criteria.
func (s Service) SearchMetrics(
	ctx context.Context, namespaceID uint, timeZoneOffset int, req request.SearchMetricsRequest,
) (*sql.Rows, int64, repositories.SearchResultMap, error) {
	rows, total, searchResult, err := s.metricRepository.SearchMetrics(ctx, namespaceID, timeZoneOffset, req)
	if err != nil {
		return nil, 0, nil, api.NewInternalError("error searching runs: %s", err)
	}
	return rows, total, searchResult, nil
}

// SearchAlignedMetrics returns the list of aligned metrics.
func (s Service) SearchAlignedMetrics(
	ctx context.Context, namespaceID uint, req *request.SearchAlignedMetricsRequest,
) (*sql.Rows, func(*sql.Rows) (*models.AlignedMetric, error), int, error) {
	// collect map of unique contexts, collect values.
	values, capacity, contextsMap := []any{}, 0, map[string]types.JSONB{}
	for _, r := range req.Runs {
		for _, t := range r.Traces {
			l := t.Slice[2]
			if l > capacity {
				capacity = l
			}
			data, err := json.Marshal(t.Context)
			if err != nil {
				return nil, nil, 0, api.NewInternalError("error serializing context: %s", err)
			}
			values, contextsMap[string(data)] = append(values, r.ID, t.Name, data, float32(l)), data
		}
	}

	contexts, err := s.metricRepository.GetContextListByContextObjects(ctx, contextsMap)
	if err != nil {
		return nil, nil, 0, api.NewInternalError("error getting context list: %s", err)
	}

	// add context ids to `values` array.
	for _, context := range contexts {
		for i := 2; i < len(values); i += 4 {
			if common.CompareJson(values[i].([]byte), context.Json) {
				values[i] = context.ID
			}
		}
	}

	rows, next, err := s.runRepository.GetAlignedMetrics(ctx, namespaceID, values, req.AlignBy)
	if err != nil {
		return nil, nil, 0, api.NewInternalError("error searching aligned run metrics: %s", err)
	}

	return rows, next, capacity, nil
}

// DeleteRun deletes requested run.
func (s Service) DeleteRun(
	ctx context.Context, namespaceID uint, req *request.DeleteRunRequest,
) error {
	run, err := s.runRepository.GetRunByNamespaceIDAndRunID(ctx, namespaceID, req.ID)
	if err != nil {
		return api.NewInternalError("error getting run by id %s: %s", req.ID, err)
	}

	if run == nil {
		return api.NewResourceDoesNotExistError("run '%s' not found", req.ID)
	}

	if err = s.runRepository.DeleteBatch(ctx, namespaceID, []string{run.ID}); err != nil {
		return api.NewInternalError("unable to delete run %q: %s", req.ID, err)
	}
	return nil
}

// UpdateRun updates requested run.
func (s Service) UpdateRun(
	ctx context.Context, namespaceID uint, req *request.UpdateRunRequest,
) error {
	run, err := s.runRepository.GetRunByNamespaceIDAndRunID(ctx, namespaceID, req.ID)
	if err != nil {
		return api.NewInternalError("error getting run by id %s: %s", req.ID, err)
	}

	if run == nil {
		return api.NewResourceDoesNotExistError("run '%s' not found", req.ID)
	}

	if req.Archived != nil {
		if *req.Archived {
			if err := s.runRepository.ArchiveBatch(ctx, namespaceID, []string{run.ID}); err != nil {
				return api.NewInternalError("error archiving run %s: %s", req.ID, err)
			}
		} else {
			if err := s.runRepository.RestoreBatch(ctx, namespaceID, []string{run.ID}); err != nil {
				return api.NewInternalError("error restoring run %s: %s", req.ID, err)
			}
		}
	}

	if req.Name != nil {
		run.Name = *req.Name
		if err := s.runRepository.Update(ctx, run); err != nil {
			return api.NewInternalError("error updating run %s: %s", req.ID, err)
		}
	}

	if req.Description != nil {
		if err := s.tagRepository.CreateRunTag(ctx, &models.Tag{
			Key:   common.DescriptionTagKey,
			Value: *req.Description,
			RunID: req.ID,
		}); err != nil {
			return api.NewInternalError("unable to create experiment tag: %s", err)
		}
	}
	return nil
}

// ProcessBatch processes runs in batch.
func (s Service) ProcessBatch(
	ctx context.Context, namespaceID uint, action string, ids []string,
) error {
	switch action {
	case BatchActionArchive:
		if err := s.runRepository.ArchiveBatch(ctx, namespaceID, ids); err != nil {
			return api.NewInternalError("error archiving runs: %s", err)
		}
	case BatchActionRestore:
		if err := s.runRepository.RestoreBatch(ctx, namespaceID, ids); err != nil {
			return api.NewInternalError("error restoring runs: %s", err)
		}
	case BatchActionDelete:
		if err := s.runRepository.DeleteBatch(ctx, namespaceID, ids); err != nil {
			return api.NewInternalError("error deleting runs: %s", err)
		}
	default:
		return eris.Errorf("unsupported batch action: %s", action)
	}
	return nil
}
