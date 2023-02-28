package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	glog "log"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
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

var DB *DbInstance = &DbInstance{}

func ConnectDB(dsn string, slowThreshold time.Duration, poolMax int, init bool, migrate bool, artifactRoot string) error {
	DB.dsn = dsn
	var sourceConn gorm.Dialector
	var replicaConn gorm.Dialector
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid database URL: %w", err)
	}
	switch u.Scheme {
	case "postgres":
		sourceConn = postgres.Open(u.String())
	case "sqlite":
		q := u.Query()
		q.Set("_case_sensitive_like", "true")
		q.Set("_mutex", "no")
		if q.Get("mode") != "memory" && !(q.Has("_journal") || q.Has("_journal_mode")) {
			q.Set("_journal", "WAL")
		}
		u.RawQuery = q.Encode()

		s, err := sql.Open(sqlite.DriverName, strings.Replace(u.String(), "sqlite://", "file:", 1))
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		DB.closers = append(DB.closers, s)
		s.SetMaxIdleConns(1)
		s.SetMaxOpenConns(4)
		s.SetConnMaxIdleTime(0)
		s.SetConnMaxLifetime(0)
		sourceConn = sqlite.Dialector{
			Conn: s,
		}

		q.Set("_query_only", "true")
		u.RawQuery = q.Encode()
		r, err := sql.Open(sqlite.DriverName, strings.Replace(u.String(), "sqlite://", "file:", 1))
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		DB.closers = append(DB.closers, r)
		replicaConn = sqlite.Dialector{
			Conn: r,
		}
	default:
		return fmt.Errorf("unsupported database scheme %s", u.Scheme)
	}

	log.Infof("Using database %s", dsn)

	dbLogLevel := logger.Warn
	if log.GetLevel() == log.DebugLevel {
		dbLogLevel = logger.Info
	}
	DB.DB, err = gorm.Open(sourceConn, &gorm.Config{
		Logger: logger.New(
			glog.New(
				log.StandardLogger().WriterLevel(log.WarnLevel),
				"",
				0,
			),
			logger.Config{
				SlowThreshold:             slowThreshold,
				LogLevel:                  dbLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if replicaConn != nil {
		DB.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: []gorm.Dialector{
					replicaConn,
				},
			}),
		)
	}

	if u.Scheme != "sqlite" {
		sqlDB, _ := DB.DB.DB()
		sqlDB.SetConnMaxIdleTime(time.Minute)
		sqlDB.SetMaxIdleConns(poolMax)
		sqlDB.SetMaxOpenConns(poolMax)
	}

	if init {
		switch u.Scheme {
		case "postgres":
			log.Info("Initializing database")
			DB.Exec("drop schema public cascade")
			DB.Exec("create schema public")
		default:
			return fmt.Errorf("unable to initialize database with scheme \"%s\"", u.Scheme)
		}
	}

	var schemaVersion AlembicVersion
	DB.Session(&gorm.Session{
		Logger: logger.Discard,
	}).First(&schemaVersion)

	if schemaVersion.Version != "97727af70f4d" {
		if !migrate {
			return fmt.Errorf("unsupported database schema version %s", schemaVersion.Version)
		}

		switch schemaVersion.Version {
		case "":
			log.Info("Migrating database to 97727af70f4d")
			tx := DB.Begin()
			if err = tx.AutoMigrate(
				&Experiment{},
				&ExperimentTag{},
				&Run{},
				&Param{},
				&Tag{},
				&Metric{},
				&LatestMetric{},
				&AlembicVersion{},
			); err != nil {
				return fmt.Errorf("error migrating database to 97727af70f4d: %w", err)
			}
			tx.Create(&AlembicVersion{
				Version: "97727af70f4d",
			})
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error setting database schema version: %s", tx.Error)
			}

		case "c48cb773bb87":
			log.Info("Migrating database to bd07f7e963c5")
			tx := DB.Begin()
			for _, table := range []any{
				&Param{},
				&Metric{},
				&LatestMetric{},
				&Tag{},
			} {
				if err := tx.Migrator().CreateIndex(table, "RunID"); err != nil {
					return fmt.Errorf("error migrating database to bd07f7e963c5: %w", err)
				}
			}
			tx.Model(&AlembicVersion{}).Where("1 = 1").Update("Version", "bd07f7e963c5")
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error setting database schema version to bd07f7e963c5: %w", err)
			}
			fallthrough

		case "bd07f7e963c5":
			log.Info("Migrating database to 0c779009ac13")
			tx := DB.Begin()
			if err := tx.Migrator().AddColumn(&Run{}, "DeletedTime"); err != nil {
				return fmt.Errorf("error migrating database to 0c779009ac13: %w", err)
			}
			tx.Model(&AlembicVersion{}).Where("1 = 1").Update("Version", "0c779009ac13")
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error setting database schema version to 0c779009ac13: %w", err)
			}
			fallthrough

		case "0c779009ac13":
			log.Info("Migrating database to cc1f77228345")
			tx := DB.Begin()
			if err := tx.Migrator().AlterColumn(&Param{}, "value"); err != nil {
				return fmt.Errorf("error migrating database to cc1f77228345: %w", err)
			}
			tx.Model(&AlembicVersion{}).Where("1 = 1").Update("Version", "cc1f77228345")
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error setting database schema version to cc1f77228345: %w", err)
			}
			fallthrough

		case "cc1f77228345":
			log.Info("Migrating database to 97727af70f4d")
			tx := DB.Begin()
			for _, column := range []string{
				"CreationTime",
				"LastUpdateTime",
			} {
				if err := tx.Migrator().AddColumn(&Experiment{}, column); err != nil {
					return fmt.Errorf("error migrating database to 97727af70f4d: %w", err)
				}
			}
			tx.Model(&AlembicVersion{}).Where("1 = 1").Update("Version", "97727af70f4d")
			tx.Commit()
			if tx.Error != nil {
				return fmt.Errorf("error setting database schema version to 97727af70f4d: %w", err)
			}

		default:
			return fmt.Errorf("unsupported database schema version %s", schemaVersion.Version)
		}
	}

	if tx := DB.First(&Experiment{}, 0); tx.Error != nil {
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
			if tx := DB.Create(&exp); tx.Error != nil {
				return fmt.Errorf("error creating default experiment: %s", tx.Error)
			}

			exp.ArtifactLocation = fmt.Sprintf("%s/%d", strings.TrimRight(artifactRoot, "/"), *exp.ID)
			if tx := DB.Model(&exp).Update("ArtifactLocation", exp.ArtifactLocation); tx.Error != nil {
				return fmt.Errorf("error updating artifact_location for experiment '%s': %s", exp.Name, tx.Error)
			}
		} else {
			return fmt.Errorf("unable to find default experiment: %s", tx.Error)
		}
	}

	return nil
}
