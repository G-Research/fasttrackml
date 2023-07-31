package fixtures

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"

	// "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ProjectFixtures represents data fixtures object.
type ProjectFixtures struct {
	baseFixtures
}

// NewProjectFixtures creates new instance of ProjectFixtures.
func NewProjectFixtures(databaseDSN string) (*ProjectFixtures, error) {
	db, err := database.ConnectDB(
		databaseDSN,
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	if err != nil {
		return nil, eris.Wrap(err, "error connection to database")
	}
	return &ProjectFixtures{
		baseFixtures: baseFixtures{db: db.DB},
	}, nil
}

func (f *ProjectFixtures) GetProject(ctx context.Context) *fiber.Map {
	return &fiber.Map{
		"name":              "FastTrackML",
		"path":              database.DB.DSN(),
		"description":       "",
		"telemetry_enabled": float64(0),
	}
}
