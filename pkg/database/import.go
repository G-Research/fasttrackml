package database

import (
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Import will copy the contents of input db to output db.
func Import(input, output *DbInstance) error {
	if err := importExperimentTree(input.DB, output.DB); err != nil {
		return eris.Wrap(err, "unable to import database")
	}
	return nil
}

// importExperimentTree will copy the contents of the experiements tree from sourceDB to destDB.
// experiment_id is renumbered.
func importExperimentTree(sourceDB, destDB *gorm.DB) error {
	// Start transaction in the destDB
	err := destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source
		rows, err := sourceDB.Model(&Experiment{}).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var item Experiment
			sourceDB.ScanRows(rows, &item)
			originalExpId := item.ID
			item.ID = nil
			if err := destTX.Clauses(
				clause.OnConflict{DoNothing: true},
				clause.Returning{Columns: []clause.Column{{Name: "experiment_id"}}},
			).Create(&item).Error; err != nil {
				return err
			}
			importExperimentTags(sourceDB, destDB, originalExpId, item.ID)
			// importRuns(sourceDB, destDB, originalExpId, item.ID)
			count++
		}
		log.Infof("Importing Experiments - found %v records", count)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func importExperimentTags(sourceDB, destDB *gorm.DB, originalID, newID *int32) error {
	// Start transaction in the destDB
	err := destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source
		rows, err := sourceDB.Model(&ExperimentTag{}).
			Where("experiment_id = ?", originalID).
			Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var item ExperimentTag
			sourceDB.ScanRows(rows, &item)
			item.ExperimentID = *newID
			if err := destTX.Clauses(
				clause.OnConflict{DoNothing: true},
				clause.Returning{Columns: []clause.Column{{Name: "experiment_id"}}},
			).Create(&item).Error; err != nil {
				return err
			}
			count++
		}
		log.Infof("Importing ExperimentTags - found %v records", count)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
