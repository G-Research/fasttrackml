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
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0c779009ac13"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_1ce8669664d2"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_5d042539be4f"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_8073e7e037e5"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_97727af70f4d"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_ac0b8b7c0014"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_bd07f7e963c5"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_cc1f77228345"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_e0d125c68d9a"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_ed364de02645"
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

	if !slices.Contains(supportedAlembicVersions, alembicVersion.Version) || schemaVersion.Version != "e0d125c68d9a" {
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
			if err := v_bd07f7e963c5.Migrate(db); err != nil {
				return fmt.Errorf("error migrating database to alembic schema bd07f7e963c5: %w", err)
			}
			fallthrough

		case "bd07f7e963c5":
			log.Info("Migrating database to alembic schema 0c779009ac13")
			if err := v_0c779009ac13.Migrate(db); err != nil {
				return fmt.Errorf("error migrating database to alembic schema 0c779009ac13: %w", err)
			}
			fallthrough

		case "0c779009ac13":
			log.Info("Migrating database to alembic schema cc1f77228345")
			if err := v_cc1f77228345.Migrate(db); err != nil {
				return fmt.Errorf("error migrating database to alembic schema cc1f77228345: %w", err)
			}
			fallthrough

		case "cc1f77228345":
			log.Info("Migrating database to alembic schema 97727af70f4d")
			if err := v_97727af70f4d.Migrate(db); err != nil {
				return fmt.Errorf("error migrating database to alembic schema 97727af70f4d: %w", err)
			}
			fallthrough

		case "97727af70f4d", "3500859a5d39", "7f2a7d5fae7d":
			switch schemaVersion.Version {
			case "":
				log.Info("Migrating database to FastTrackML schema ac0b8b7c0014")
				if err := v_ac0b8b7c0014.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema ac0b8b7c0014: %w", err)
				}
				fallthrough

			case "ac0b8b7c0014":
				log.Info("Migrating database to FastTrackML schema 8073e7e037e5")
				if err := v_8073e7e037e5.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema 8073e7e037e5: %w", err)
				}
				fallthrough

			case "8073e7e037e5":
				log.Info("Migrating database to FastTrackML schema ed364de02645")
				if err := v_ed364de02645.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema ed364de02645: %w", err)
				}
				fallthrough

			case "ed364de02645":
				log.Info("Migrating database to FastTrackML schema 1ce8669664d2")
				if err := v_1ce8669664d2.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema 1ce8669664d2: %w", err)
				}
				fallthrough

			case "1ce8669664d2":
				log.Info("Migrating database to FastTrackML schema 5d042539be4f")
				if err := v_5d042539be4f.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema 5d042539be4f: %w", err)
				}
				fallthrough

			case "5d042539be4f":
				log.Info("Migrating database to FastTrackML schema e0d125c68d9a")
				if err := v_e0d125c68d9a.Migrate(db); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema e0d125c68d9a: %w", err)
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
				Version: "e0d125c68d9a",
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
				Description:         "Default",
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
