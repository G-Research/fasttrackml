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

// DBProvider is the interface to access the DB.
type DBProvider interface {
	Db() *gorm.DB
	Dsn() string
	Close() error
	Reset() error
}

// DB is a global gorm.DB reference
var DB *gorm.DB

// DBInstance is the base concrete type for DbProvider.
type DBInstance struct {
	*gorm.DB
	dsn     string
	closers []io.Closer
}

// Close will invoke the closers.
func (db *DBInstance) Close() error {
	for _, c := range db.closers {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Dsn will return the dsn string.
func (db *DBInstance) Dsn() string {
	return db.dsn
}

// Db will return the gorm DB.
func (db *DBInstance) Db() *gorm.DB {
	return db.DB
}

// createDefaultExperiment will create the default experiment if needed.
func createDefaultExperiment(artifactRoot string, db DBProvider) error {
	if tx := db.Db().First(&Experiment{}, 0); tx.Error != nil {
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
			if tx := db.Db().Create(&exp); tx.Error != nil {
				return fmt.Errorf("error creating default experiment: %s", tx.Error)
			}

			exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(artifactRoot, "/"), *exp.ID)
			if tx := db.Db().Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
				return fmt.Errorf("error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
			}
		} else {
			return fmt.Errorf("unable to find default experiment: %s", tx.Error)
		}
	}
	return nil
}
