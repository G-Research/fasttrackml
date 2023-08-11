package database

import (
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Import will copy the contents of input db to output db.
func Import(input, output *DbInstance) error {
	in := input.DB
	tables := []string{
		"experiments",
		"experiment_tags",
		"runs",
		"tags",
		"params",
		"metrics",
		"latest_metrics",
		"apps",
		"dashboards",
	}
	if err := output.DB.Transaction(func(out *gorm.DB) error {
		for _, table := range tables {
			if err := importTable(in, out, table); err != nil {
				return eris.Wrapf(err, "error importing table %s", table)
			}
		}
		return nil
	}); err != nil {
		return eris.Wrapf(err, "error importing database")
	}

	return nil
}

// importTable will copy the contents of one table (model) from sourceDB to destDB.
func importTable(sourceDB, destDB *gorm.DB, table string) error {
	// Start transaction in the destDB
	err := destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source database
		rows, err := sourceDB.Table(table).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var item map[string]any
			sourceDB.ScanRows(rows, &item)
			if err := destTX.
				Table(table).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&item).Error; err != nil {
				return err
			}
			count++
		}
		log.Infof("Importing %s - found %v records", table, count)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
