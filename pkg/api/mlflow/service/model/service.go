package model

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// Service provides service layer to work with `model` business logic.
type Service struct{}

// NewService creates new Service instance.
func NewService() *Service {
	return &Service{}
}

func (s Service) SearchModelVersions(ctx context.Context) (any, error) {
	return fiber.Map{
		"model_versions": []any{},
	}, nil
}

func (s Service) SearchRegisteredModels(ctx context.Context) (any, error) {
	return fiber.Map{
		"registered_models": []any{},
	}, nil
}
