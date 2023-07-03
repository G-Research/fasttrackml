package fixtures

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// AppFixtures represents data fixtures object.
type AppFixtures struct {
	baseFixtures
	*database.DbInstance
}

// NewAppFixtures creates new instance of AppFixtures.
func NewAppFixtures(databaseDSN string) (*AppFixtures, error) {
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
	return &AppFixtures{
		baseFixtures: baseFixtures{db: db.DB},
		DbInstance:   db,
	}, nil
}

// CreateTestApp creates a new test App.
func (f AppFixtures) CreateTestApp(
	ctx context.Context, app *models.App,
) (*models.App, error) {
	if err := f.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test app")
	}
	return app, nil
}

// CreateTestApps creates some num apps belonging to the experiment
func (f AppFixtures) CreateTestApps(
	ctx context.Context, num int,
) ([]*models.App, error) {
	var apps []*models.App
	// create apps for the experiment
	for i := 0; i < num; i++ {
		app := &models.App{
			ID:    uuid.New(),
			Type:  "mpi",
			State: models.AppState{},
		}
		app, err := f.CreateTestApp(context.Background(), app)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// GetTestApps fetches all apps for an experiment
func (f AppFixtures) GetTestApps(
	ctx context.Context) ([]models.App, error) {
	apps := []models.App{}
	if err := f.db.WithContext(ctx).
		Order("start_time desc").
		Find(&apps).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting `app` entities")
	}
	return apps, nil
}
