package database

import (
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Import will copy the contents of input db to output db.
func Import(input, output *DbInstance, dryRun bool) error {
	in := input.DB
	if err := output.DB.Transaction(func(out *gorm.DB) error {
		if err := importTable(in, out, dryRun, Experiment{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, ExperimentTag{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, Run{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, Tag{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, Metric{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, LatestMetric{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, Param{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, Dashboard{}); err != nil {
			return err
		}
		if err := importTable(in, out, dryRun, App{}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return eris.Wrapf(err, "error importing table data")
	}

	return nil
}

// importTable will copy the contents of one table (model) from sourceDB to destDB.
func importTable[T any](sourceDB, destDB *gorm.DB, dryRun bool, model T) error {
	// Start transaction in the destDB
	err := destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source database
		rows, err := sourceDB.Model(&model).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var item T
			sourceDB.ScanRows(rows, &item)
			if !dryRun {
				if err := destTX.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
					return err
				}
			}
			count++
		}
		log.Infof("Importing %T - found %v records", model, count)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
