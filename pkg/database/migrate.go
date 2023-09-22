package database

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var supportedAlembicVersions = []string{
	"97727af70f4d",
	"3500859a5d39",
	"7f2a7d5fae7d",
}

func checkAndMigrate(migrate bool, dbProvider DBProvider) error {
	db := dbProvider.GormDB()
	var alembicVersion AlembicVersion
	var schemaVersion SchemaVersion
	{
		tx := db.Session(&gorm.Session{
			Logger: logger.Discard,
		})
		tx.First(&alembicVersion)
		tx.First(&schemaVersion)
	}

	if !slices.Contains(supportedAlembicVersions, alembicVersion.Version) || schemaVersion.Version != "5d042539be4f" {
		if !migrate && alembicVersion.Version != "" {
			return fmt.Errorf("unsupported database schema versions alembic %s, FastTrackML %s", alembicVersion.Version, schemaVersion.Version)
		}

		runWithoutForeignKeyIfNeeded := func(fn func() error) error { return fn() }
		switch db.Dialector.Name() {
		case "sqlite":
			migrator := db.Migrator().(sqlite.Migrator)
			runWithoutForeignKeyIfNeeded = migrator.RunWithoutForeignKey
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
				log.Info("Migrating database to FastTrackML schema ac0b8b7c0014")
				if err := db.Transaction(func(tx *gorm.DB) error {
					for _, column := range []struct {
						dst   any
						field string
					}{
						{&Run{}, "RowNum"},
						{&Metric{}, "Iter"},
						{&LatestMetric{}, "LastIter"},
					} {
						if err := tx.Migrator().AddColumn(column.dst, column.field); err != nil {
							return err
						}
					}
					if err := tx.Exec(
						"UPDATE runs" +
							"  SET row_num = rows.row_num" +
							"  FROM (" +
							"    SELECT run_uuid, ROW_NUMBER() OVER (ORDER BY start_time, run_uuid DESC) - 1 AS row_num" +
							"    FROM runs" +
							"  ) AS rows" +
							"  WHERE runs.run_uuid = rows.run_uuid").
						Error; err != nil {
						return err
					}
					if err := tx.Exec(
						"UPDATE metrics" +
							"  SET iter = iters.iter" +
							"  FROM (" +
							"    SELECT ROW_NUMBER() OVER (PARTITION BY run_uuid, key ORDER BY timestamp, step, value) - 1 AS iter," +
							"      run_uuid, key, timestamp, step, value" +
							"    FROM metrics" +
							"  ) AS iters" +
							"  WHERE" +
							"    (metrics.run_uuid, metrics.key, metrics.timestamp, metrics.step, metrics.value) =" +
							"    (iters.run_uuid, iters.key, iters.timestamp, iters.step, iters.value)").
						Error; err != nil {
						return err
					}
					if err := tx.Exec(
						"UPDATE latest_metrics" +
							"  SET last_iter = metrics.last_iter" +
							"  FROM (" +
							"    SELECT run_uuid, key, MAX(iter) AS last_iter" +
							"    FROM metrics" +
							"    GROUP BY run_uuid, key" +
							"  ) AS metrics" +
							"  WHERE" +
							"    (latest_metrics.run_uuid, latest_metrics.key) =" +
							"    (metrics.run_uuid, metrics.key)").
						Error; err != nil {
						return err
					}
					if err := tx.AutoMigrate(&SchemaVersion{}); err != nil {
						return err
					}
					return tx.Create(&SchemaVersion{
						Version: "ac0b8b7c0014",
					}).Error
				}); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema ac0b8b7c0014: %w", err)
				}
				fallthrough

			case "ac0b8b7c0014":
				log.Info("Migrating database to FastTrackML schema 8073e7e037e5")
				if err := db.Transaction(func(tx *gorm.DB) error {
					if err := tx.AutoMigrate(
						&Dashboard{},
						&App{},
					); err != nil {
						return err
					}
					return tx.Model(&SchemaVersion{}).
						Where("1 = 1").
						Update("Version", "8073e7e037e5").
						Error
				}); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema 8073e7e037e5: %w", err)
				}
				fallthrough

			case "8073e7e037e5":
				log.Info("Migrating database to FastTrackML schema ed364de02645")
				if err := db.Transaction(func(tx *gorm.DB) error {
					if err := tx.Migrator().CreateIndex(&Run{}, "RowNum"); err != nil {
						return err
					}
					if err := tx.Migrator().CreateIndex(&Metric{}, "Iter"); err != nil {
						return err
					}
					return tx.Model(&SchemaVersion{}).
						Where("1 = 1").
						Update("Version", "ed364de02645").
						Error
				}); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema ed364de02645: %w", err)
				}
				fallthrough

			case "ed364de02645":
				log.Info("Migrating database to FastTrackML schema 1ce8669664d2")
				if err := db.Transaction(func(tx *gorm.DB) error {
					constraints := []string{"Params", "Tags", "Metrics", "LatestMetrics"}
					for _, constraint := range constraints {
						switch tx.Dialector.Name() {
						case "sqlite":
							// SQLite tables need to be recreated to add or remove constraints.
							// By not dropping the constraint, we can avoid having to recreate the table twice.
						case "postgres":
							// Existing MLFlow Postgres databases have foreign key constraints
							// with their own names. We need to drop them before we can add our own.
							table := tx.NamingStrategy.TableName(constraint)
							fk := fmt.Sprintf("%s_run_uuid_fkey", table)
							if tx.Migrator().HasConstraint(table, fk) {
								if err := tx.Migrator().DropConstraint(table, fk); err != nil {
									return err
								}
							} else {
								if err := tx.Migrator().DropConstraint(&Run{}, constraint); err != nil {
									return err
								}
							}
						default:
							return fmt.Errorf("unsupported database dialect %s", tx.Dialector.Name())
						}

						if err := tx.Migrator().CreateConstraint(&Run{}, constraint); err != nil {
							return err
						}
					}
					return tx.Model(&SchemaVersion{}).
						Where("1 = 1").
						Update("Version", "1ce8669664d2").
						Error
				}); err != nil {
					return fmt.Errorf("error migrating database to FastTrackML schema 1ce8669664d2: %w", err)
				}
				fallthrough

			case "1ce8669664d2":
				log.Info("Migrating database to FastTrackML schema 5d042539be4f")
				// We need to run this migration without foreign key constraints to avoid
				// the cascading delete to kick in and delete all the run data.
				if err := runWithoutForeignKeyIfNeeded(func() error {
					if err := db.Transaction(func(tx *gorm.DB) error {
						elems := []struct {
							Table      string
							Constraint string
						}{
							{"experiment_tags", "Tags"},
							{"runs", "Runs"},
						}
						for _, e := range elems {
							switch tx.Dialector.Name() {
							case "sqlite":
								// SQLite tables need to be recreated to add or remove constraints.
								// By not dropping the constraint, we can avoid having to recreate the table twice.
							case "postgres":
								// Existing MLFlow Postgres databases have foreign key constraints with their own names.
								// We need to drop them before we can add our own.
								fk := fmt.Sprintf("%s_experiment_id_fkey", e.Table)
								if tx.Migrator().HasConstraint(e.Table, fk) {
									if err := tx.Migrator().DropConstraint(e.Table, fk); err != nil {
										return err
									}
								} else {
									if err := tx.Migrator().DropConstraint(&Experiment{}, e.Constraint); err != nil {
										return err
									}
								}
							default:
								return fmt.Errorf("unsupported database dialect %s", tx.Dialector.Name())
							}

							if err := tx.Migrator().CreateConstraint(&Experiment{}, e.Constraint); err != nil {
								return err
							}
						}
						return tx.Model(&SchemaVersion{}).
							Where("1 = 1").
							Update("Version", "5d042539be4f").
							Error
					}); err != nil {
						return fmt.Errorf("error migrating database to FastTrackML schema 5d042539be4f: %w", err)
					}
					return nil
				}); err != nil {
					return err
				}

			default:
				return fmt.Errorf("unsupported database FastTrackML schema version %s", schemaVersion.Version)
			}

			log.Info("Database migration done")

		case "":
			log.Info("Initializing database")
			tx := db.Begin()
			if err := tx.AutoMigrate(
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
				Version: "5d042539be4f",
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
