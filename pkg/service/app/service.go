package app

import (
	"fmt"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// Service provides service layer to work with `model` business logic.
type Service struct{}

// NewService creates new Service instance.
func NewService() *Service {
	return &Service{}
}

func (svc Service) GetApps() ([]database.App, error) {
	var apps []database.App
	if err := database.DB.
		Where("NOT is_archived").
		Find(&apps).
		Error; err != nil {
		return nil, fmt.Errorf("error fetching apps: %w", err)
	}
	return apps, nil
}
