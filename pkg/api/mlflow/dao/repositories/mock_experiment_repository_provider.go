// Code generated by mockery v2.34.0. DO NOT EDIT.

package repositories

import (
	context "context"

	models "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	mock "github.com/stretchr/testify/mock"
)

// MockExperimentRepositoryProvider is an autogenerated mock type for the ExperimentRepositoryProvider type
type MockExperimentRepositoryProvider struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, experiment
func (_m *MockExperimentRepositoryProvider) Create(ctx context.Context, experiment *models.Experiment) error {
	ret := _m.Called(ctx, experiment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Experiment) error); ok {
		r0 = rf(ctx, experiment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, experiment
func (_m *MockExperimentRepositoryProvider) Delete(ctx context.Context, experiment *models.Experiment) error {
	ret := _m.Called(ctx, experiment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Experiment) error); ok {
		r0 = rf(ctx, experiment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteBatch provides a mock function with given fields: ctx, ids
func (_m *MockExperimentRepositoryProvider) DeleteBatch(ctx context.Context, ids []*int32) error {
	ret := _m.Called(ctx, ids)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*int32) error); ok {
		r0 = rf(ctx, ids)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByNamespaceIDAndExperimentID provides a mock function with given fields: ctx, namespaceID, experimentID
func (_m *MockExperimentRepositoryProvider) GetByNamespaceIDAndExperimentID(ctx context.Context, namespaceID uint, experimentID int32) (*models.Experiment, error) {
	ret := _m.Called(ctx, namespaceID, experimentID)

	var r0 *models.Experiment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uint, int32) (*models.Experiment, error)); ok {
		return rf(ctx, namespaceID, experimentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uint, int32) *models.Experiment); ok {
		r0 = rf(ctx, namespaceID, experimentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Experiment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint, int32) error); ok {
		r1 = rf(ctx, namespaceID, experimentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByNamespaceIDAndName provides a mock function with given fields: ctx, namespaceID, name
func (_m *MockExperimentRepositoryProvider) GetByNamespaceIDAndName(ctx context.Context, namespaceID uint, name string) (*models.Experiment, error) {
	ret := _m.Called(ctx, namespaceID, name)

	var r0 *models.Experiment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uint, string) (*models.Experiment, error)); ok {
		return rf(ctx, namespaceID, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uint, string) *models.Experiment); ok {
		r0 = rf(ctx, namespaceID, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Experiment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint, string) error); ok {
		r1 = rf(ctx, namespaceID, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, experiment
func (_m *MockExperimentRepositoryProvider) Update(ctx context.Context, experiment *models.Experiment) error {
	ret := _m.Called(ctx, experiment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Experiment) error); ok {
		r0 = rf(ctx, experiment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockExperimentRepositoryProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockExperimentRepositoryProvider creates a new instance of MockExperimentRepositoryProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockExperimentRepositoryProvider(t mockConstructorTestingTNewMockExperimentRepositoryProvider) *MockExperimentRepositoryProvider {
	mock := &MockExperimentRepositoryProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
