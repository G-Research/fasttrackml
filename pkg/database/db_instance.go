package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DbInstance struct {
	*gorm.DB
	dsn     string
	closers []io.Closer
}

func (db *DbInstance) Close() error {
	for _, c := range db.closers {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DbInstance) DSN() string {
	return db.dsn
}

// DB is a global db instance.
var DB *DbInstance = &DbInstance{}

func (db *DbInstance) reset() error {
	switch db.Dialector.Name() {
	case "postgres":
		log.Info("Resetting database schema")
		db.Exec("drop schema public cascade")
		db.Exec("create schema public")
	default:
		return fmt.Errorf("unable to reset database with backend \"%s\"", db.Dialector.Name())
	}
	return nil
}

func (db *DbInstance) createDefaultExperiment(artifactRoot string) error {
	if tx := db.First(&Experiment{}, 0); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			log.Info("Creating default experiment")
			var id int32 = 0
			ts := time.Now().UTC().UnixMilli()
			exp := Experiment{
				ID:             &id,
				Name:           "Default",
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
			if tx := db.Create(&exp); tx.Error != nil {
				return fmt.Errorf("error creating default experiment: %s", tx.Error)
			}

			exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(artifactRoot, "/"), *exp.ID)
			if tx := db.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
				return fmt.Errorf("error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
			}
		} else {
			return fmt.Errorf("unable to find default experiment: %s", tx.Error)
		}
	}
	return nil
}
