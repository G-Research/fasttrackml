// Code generated by mockery v2.34.0. DO NOT EDIT.

package repositories

import (
	context "context"

	gorm "gorm.io/gorm"

	mock "github.com/stretchr/testify/mock"

	models "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// MockLogRepositoryProvider is an autogenerated mock type for the LogRepositoryProvider type
type MockLogRepositoryProvider struct {
	mock.Mock
}

// GetDB provides a mock function with given fields:
func (_m *MockLogRepositoryProvider) GetDB() *gorm.DB {
	ret := _m.Called()

	var r0 *gorm.DB
	if rf, ok := ret.Get(0).(func() *gorm.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	return r0
}

// SaveLog provides a mock function with given fields: ctx, log
func (_m *MockLogRepositoryProvider) SaveLog(ctx context.Context, log *models.Log) error {
	ret := _m.Called(ctx, log)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Log) error); ok {
		r0 = rf(ctx, log)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockLogRepositoryProvider creates a new instance of MockLogRepositoryProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockLogRepositoryProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLogRepositoryProvider {
	mock := &MockLogRepositoryProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
