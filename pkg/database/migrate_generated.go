// Code generated by 'make migrations-rebuild'; DO NOT EDIT.
package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0001"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0002"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0003"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0004"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0005"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0006"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0007"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0008"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0009"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0010"
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0011"
)

func currentVersion() string {
	return v_0011.Version
}

func generatedMigrations(db *gorm.DB, schemaVersion string) error {
	switch schemaVersion {
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
		fallthrough

	case v_0007.Version:
		log.Infof("Migrating database to FastTrackML schema %s", v_0008.Version)
		if err := v_0008.Migrate(db); err != nil {
			return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0008.Version, err)
		}
		fallthrough

	case v_0008.Version:
		log.Infof("Migrating database to FastTrackML schema %s", v_0009.Version)
		if err := v_0009.Migrate(db); err != nil {
			return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0009.Version, err)
		}
		fallthrough

	case v_0009.Version:
		log.Infof("Migrating database to FastTrackML schema %s", v_0010.Version)
		if err := v_0010.Migrate(db); err != nil {
			return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0010.Version, err)
		}
		fallthrough

	case v_0010.Version:
		log.Infof("Migrating database to FastTrackML schema %s", v_0011.Version)
		if err := v_0011.Migrate(db); err != nil {
			return fmt.Errorf("error migrating database to FastTrackML schema %s: %w", v_0011.Version, err)
		}

	default:
		return fmt.Errorf("unsupported database FastTrackML schema version %s", schemaVersion)
	}
	return nil
}
