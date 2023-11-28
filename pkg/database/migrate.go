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
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0001"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0002"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0003"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0004"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0005"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0006"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0007"
)

var supportedAlembicVersions = []string{
	"97727af70f4d",
	"3500859a5d39",
	"7f2a7d5fae7d",
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

	if !slices.Contains(supportedAlembicVersions, alembicVersion.Version) || schemaVersion.Version != v_0007.Version {
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
					&Param{},
					&Metric{},
					&LatestMetric{},
					&Tag{},
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
				if err := tx.Migrator().AddColumn(&Run{}, "DeletedTime"); err != nil {
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
				if err := tx.Migrator().AlterColumn(&Param{}, "value"); err != nil {
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
					if err := tx.Migrator().AddColumn(&Experiment{}, column); err != nil {
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

		case "97727af70f4d", "3500859a5d39", "7f2a7d5fae7d":
			switch schemaVersion.Version {
			case "":
				log.Infof("Migrating database to FastTrackML schema %s", v_0001.Version)
				if err := v_0001.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0001.Version, err)
				}
				fallthrough

			case v_0001.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0002.Version)
				if err := v_0002.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0002.Version, err)
				}
				fallthrough

			case v_0002.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0003.Version)
				if err := v_0003.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0003.Version, err)
				}
				fallthrough

			case v_0003.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0004.Version)
				if err := v_0004.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0004.Version, err)
				}
				fallthrough

			case v_0004.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0005.Version)
				if err := v_0005.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0005.Version, err)
				}
				fallthrough

			case v_0005.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0006.Version)
				if err := v_0006.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0006.Version, err)
				}
				fallthrough

			case v_0006.Version:
				log.Infof("Migrating database to FastTrackML schema %s", v_0007.Version)
				if err := v_0007.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0007.Version, err)
				}

			default:
				return fmt.Errorf("unsupported database FastTrackML schema version %s", schemaVersion.Version)
			}

			log.Info("Database migration done")

		case "":
			log.Info("Initializing database")
			tx := db.Begin()
			if err := tx.AutoMigrate(
				&Namespace{},
				&Experiment{},
				&ExperimentTag{},
				&Run{},
				&Param{},
				&Tag{},
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
				Version: v_0007.Version,
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
	if tx := db.First(&Namespace{
		Code: "default",
	}); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			log.Info("Creating default namespace")
			var exp int32 = 0
			ns := Namespace{
				Code:                "default",
				Description:         "Default namespace",
				DefaultExperimentID: &exp,
			}
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&ns).Error; err != nil {
					return err
				}
				if err := tx.Model(&Experiment{}).
					Where("namespace_id IS NULL").
					Update("namespace_id", ns.ID).
					Error; err != nil {
					return fmt.Errorf("error updating experiments: %s", err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("error creating default namespace: %s", err)
			}
		} else {
			return fmt.Errorf("unable to find default namespace: %s", tx.Error)
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
					ID:             common.GetPointer(int32(0)),
					Name:           "Default",
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
