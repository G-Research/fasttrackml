package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
)

// AppRepositoryProvider provides an interface to work with models.Tag entity.
type AppRepositoryProvider interface {
	BaseRepositoryProvider
	// CreateExperimentTag creates new models.ExperimentTag entity connected to models.Experiment.
	CreateExperimentTag(ctx context.Context, experimentTag *models.ExperimentTag) error
	// CreateRunTagWithTransaction creates new models.Tag entity connected to models.Run.
	CreateRunTagWithTransaction(ctx context.Context, tx *gorm.DB, runID, key, value string) error
	// GetByRunIDAndKey returns models.Tag by provided RunID and Tag Key.
	GetByRunIDAndKey(ctx context.Context, runID, key string) (*models.Tag, error)
	// Delete deletes existing models.Tag entity.
	Delete(ctx context.Context, tag *models.Tag) error
}

// AppRepository repository to work with models.Tag entity.
type AppRepository struct {
	BaseRepository
}

// NewAppRepository creates repository to work with models.Tag entity.
func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{
		BaseRepository{
			db: db,
		},
	}
}

func (r AppRepository) GetApps() ([]database.App, error) {
	var apps []database.App
	err := database.DB.Where("NOT is_archived").Find(&apps).Error
	if err != nil {
		return apps, fmt.Errorf("error fetching apps: %w", err)
	}

	return apps, nil
}

func (r AppRepository) CreateApp(app database.App) (database.App, error) {
	err := r.BaseRepository.DB.Create(&app).Error
	if err != nil {
		return app, fmt.Sprintf("error inserting app: %s", err)
	}
	return app, nil
}


func (r AppRepository) GetApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	return c.JSON(app)
}

func (r AppRepository) UpdateApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var a struct {
		Type  string
		State database.AppState
	}

	if err := c.BodyParser(&a); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&app).
		Updates(database.App{
			Type:  a.Type,
			State: a.State,
		}).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating app %q: %s", p.ID, err))
	}

	return c.JSON(app)
}

func (r AppRepository) DeleteApp(c *fiber.Ctx) error {
	p := struct {
		ID uuid.UUID `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	app := database.App{
		Base: database.Base{
			ID: p.ID,
		},
	}
	if err := database.DB.
		Select("ID").
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", p.ID, err))
	}

	if err := database.DB.
		Model(&app).
		Update("IsArchived", true).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", p.ID, err))
	}

	return c.Status(200).JSON(nil)
}

