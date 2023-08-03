package database

import (
	"fmt"

	"github.com/rotisserie/eris"
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

	fmt.Println("Data transfer complete")
	return nil
}

// importTable will copy the contents of one table (model) from sourceDB to destDB.
func importTable[T any](sourceDB, destDB *gorm.DB, dryRun bool, model T) error {

	// Query data from the source database
	var sourceData []T
	fmt.Printf("Importing %T\n", sourceData)
	tx := sourceDB.Model(&model).Find(&sourceData)
	if tx.Error != nil {
		return tx.Error
	}

	fmt.Printf("Transferring %d records (dry run? %v)\n\n", len(sourceData), dryRun)

	// Transfer data to the destination database in bulk with collision handling
	var destModels []T
	for _, item := range sourceData {

		found, err := findCollision(destDB, item)
		if err != nil {
			return err
		}
		if  found == true {
		 	// Handle the collision with a warning (print to console in this example)
		 	fmt.Printf("Warning: Skipping record (type/values) '%T/%v', a record with the same ID already exists.\n", item, item )
		} else {
			destModels = append(destModels, item)
		}
	}

	// Perform bulk insert of non-colliding records
	if len(destModels) > 0 && !dryRun {
		destDB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(destModels, len(destModels))
	}
	return nil
}

// findCollision will return true when the sourceItem appears to already exist in the
// destination DB
func findCollision(destDB *gorm.DB, sourceItem any) (bool, error) {
	switch sourceItem.(type) {
	case Experiment:
		typedItem := sourceItem.(Experiment)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"name = ?",
			typedItem.Name,
		).Count(&c)
		return c > 0, tx.Error
	case ExperimentTag:
		typedItem := sourceItem.(ExperimentTag)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"experiment_id = ? AND key = ?",
			typedItem.ExperimentID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Run:
		typedItem := sourceItem.(Run)
		c := int64(0)
		tx := destDB.Model(typedItem).Where("run_uuid = ?",
			typedItem.ID,
		).Count(&c)
		return c > 0, tx.Error
	case Param:
		typedItem := sourceItem.(Param)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Tag:
		typedItem := sourceItem.(Tag)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Metric:
		typedItem := sourceItem.(Metric)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case LatestMetric:
		typedItem := sourceItem.(LatestMetric)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Dashboard:
		typedItem := sourceItem.(Dashboard)
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"name = ? AND app_id = ?",
			typedItem.Name,
			typedItem.AppID,
		).Count(&c)
		return c > 0, tx.Error
	case App:
		typedItem := sourceItem.(App)
		c := int64(0)
		tx := destDB.Model(typedItem).Where("id = ? ", typedItem.ID).Count(&c)
		return c > 0, tx.Error
	default:
		return false, eris.New("Could not determine source item type")
	}
}
