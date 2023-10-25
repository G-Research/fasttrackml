package v_bd07f7e963c5

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
	})
}
