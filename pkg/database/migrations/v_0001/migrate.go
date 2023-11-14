package v_0001

import (
	"gorm.io/gorm"
)

const Version = "ac0b8b7c0014"

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
			Version: Version,
		}).Error
	})
}
