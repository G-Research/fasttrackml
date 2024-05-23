package database

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/types"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0001"
)

var supportedAlembicVersions = []string{
	"97727af70f4d",
	"3500859a5d39",
	"7f2a7d5fae7d",
	"2d6e25af4d3e",
	"acf3f17fdcc7",
	"867495a8f9d4",
}

// CheckAndMigrateDB makes database migration.
// nolint:gocyclo
func CheckAndMigrateDB(migrate bool, db *gorm.DB) error {
	var alembicVersion AlembicVersion
	var schemaVersion SchemaVersion
	{
		tx := db.Session(&gorm.Session{
			Logger: logger.Discard,
		})
		tx.First(&alembicVersion)
		tx.First(&schemaVersion)
	}

	if !slices.Contains(supportedAlembicVersions, alembicVersion.Version) || schemaVersion.Version != currentVersion() {
		if !migrate && alembicVersion.Version != "" {
			return fmt.Errorf(
				"unsupported database schema versions alembic %s, FastTrackML %s",
				alembicVersion.Version,
				schemaVersion.Version,
			)
		}

		switch alembicVersion.Version {
		case "c48cb773bb87":
			log.Info("Migrating database to alembic schema bd07f7e963c5")
			if err := db.Transaction(func(tx *gorm.DB) error {
				for _, table := range []any{
					&v_0001.Param{},
					&v_0001.Metric{},
					&v_0001.LatestMetric{},
					&v_0001.Tag{},
				} {
					if err := tx.Migrator().CreateIndex(table, "RunID"); err != nil {
						return err
					}
				}
				return tx.Model(&AlembicVersion{}).
					Where("1 = 1").
					Update("Version", "bd07f7e963c5").
					Error
			}); err != nil {
				return fmt.Errorf("error migrating database to alembic schema bd07f7e963c5: %w", err)
			}
			fallthrough

		case "bd07f7e963c5":
			log.Info("Migrating database to alembic schema 0c779009ac13")
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Migrator().AddColumn(&v_0001.Run{}, "DeletedTime"); err != nil {
					return err
				}
				return tx.Model(&AlembicVersion{}).
					Where("1 = 1").
					Update("Version", "0c779009ac13").
					Error
			}); err != nil {
				return fmt.Errorf("error migrating database to alembic schema 0c779009ac13: %w", err)
			}
			fallthrough

		case "0c779009ac13":
			log.Info("Migrating database to alembic schema cc1f77228345")
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Migrator().AlterColumn(&v_0001.Param{}, "value"); err != nil {
					return err
				}
				return tx.Model(&AlembicVersion{}).
					Where("1 = 1").
					Update("Version", "cc1f77228345").
					Error
			}); err != nil {
				return fmt.Errorf("error migrating database to alembic schema cc1f77228345: %w", err)
			}
			fallthrough

		case "cc1f77228345":
			log.Info("Migrating database to alembic schema 97727af70f4d")
			if err := db.Transaction(func(tx *gorm.DB) error {
				for _, column := range []string{
					"CreationTime",
					"LastUpdateTime",
				} {
					if err := tx.Migrator().AddColumn(&v_0001.Experiment{}, column); err != nil {
						return err
					}
				}
				return tx.Model(&AlembicVersion{}).
					Where("1 = 1").
					Update("Version", "97727af70f4d").
					Error
			}); err != nil {
				return fmt.Errorf("error migrating database to alembic schema 97727af70f4d: %w", err)
			}
			fallthrough

		case "97727af70f4d", "3500859a5d39", "7f2a7d5fae7d", "2d6e25af4d3e", "acf3f17fdcc7", "867495a8f9d4":
			// run the FML migrations generated by `make migrations-rebuild`
			if err := generatedMigrations(db, schemaVersion.Version); err != nil {
				return fmt.Errorf("error running generated migrations: %w", err)
			}
		case "":
			log.Info("Initializing database")
			tx := db.Begin()
			if err := tx.AutoMigrate(
				&Role{},
				&Namespace{},
				&RoleNamespace{},
				&Experiment{},
				&ExperimentTag{},
				&Run{},
				&Param{},
				&Tag{},
				&Context{},
				&Metric{},
				&LatestMetric{},
				&AlembicVersion{},
				&Dashboard{},
				&App{},
				&SchemaVersion{},
			); err != nil {
				return fmt.Errorf("error initializing database: %w", err)
			}
			tx.Create(&AlembicVersion{
				Version: "97727af70f4d",
			})
			tx.Create(&SchemaVersion{
				Version: currentVersion(),
			})
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error initializing database: %w", tx.Error)
			}

		default:
			return fmt.Errorf("unsupported database alembic schema version %s", alembicVersion.Version)
		}
	}

	return nil
}

// CreateDefaultNamespace creates the default namespace if it doesn't exist.
func CreateDefaultNamespace(db *gorm.DB) error {
	if err := db.First(&Namespace{
		Code: "default",
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Creating default namespace")
			var exp int32 = 0
			ns := Namespace{
				Code:                models.DefaultNamespaceCode,
				Description:         "Default namespace",
				DefaultExperimentID: &exp,
			}
			if err := db.Create(&ns).Error; err != nil {
				return fmt.Errorf("error creating default namespace: %s", err)
			}
		} else {
			return fmt.Errorf("unable to find default namespace: %s", err)
		}
	}
	return nil
}

// CreateDefaultExperiment creates the default experiment if it doesn't exist.
func CreateDefaultExperiment(db *gorm.DB, defaultArtifactRoot string) error {
	if err := db.First(&Experiment{}, 0).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Creating default experiment")
			ns := Namespace{Code: "default"}
			if err = db.Find(&ns).Error; err != nil {
				return fmt.Errorf("error finding default namespace: %s", err)
			}

			if err := db.Transaction(func(tx *gorm.DB) error {
				ts := time.Now().UTC().UnixMilli()
				exp := Experiment{
					ID:             common.GetPointer(models.DefaultExperimentID),
					Name:           models.DefaultExperimentName,
					NamespaceID:    ns.ID,
					LifecycleStage: LifecycleStageActive,
					CreationTime: sql.NullInt64{
						Int64: ts,
						Valid: true,
					},
					LastUpdateTime: sql.NullInt64{
						Int64: ts,
						Valid: true,
					},
				}

				if err := tx.Create(&exp).Error; err != nil {
					return err
				}

				exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(defaultArtifactRoot, "/"), *exp.ID)
				if err := tx.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation).Error; err != nil {
					return fmt.Errorf("error updating artifact_location for experiment '%s': %s", exp.Name, err)
				}

				return nil
			}); err != nil {
				return fmt.Errorf("error creating default experiment: %s", err)
			}
		} else {
			return fmt.Errorf("unable to find default experiment: %s", err)
		}
	}
	return nil
}

// CreateDefaultMetricContext creates the default metric context if it doesn't exist.
func CreateDefaultMetricContext(db *gorm.DB) error {
	defaultContext := Context{Json: types.JSONB("{}")}
	if err := db.First(&defaultContext).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("Creating default context")
			if err := db.Create(&defaultContext).Error; err != nil {
				return fmt.Errorf("error creating default context: %s", err)
			}
		} else {
			return fmt.Errorf("unable to find default context: %s", err)
		}
	}
	log.Debugf("default metric context: %v", defaultContext)
	return nil
}
