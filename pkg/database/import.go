package database

import (
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	// Query data from the source database
	log.Infof("Importing %T", model)
	rows, err := sourceDB.Model(&model).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	log.Infof("Transferring records (dry run? %v)", dryRun)

	// Stream data to the destination database in bulk with collision handling
	for rows.Next() {
		var item T
		sourceDB.ScanRows(rows, &item)
		found, err := findCollision(destDB, item)
		if err != nil {
			return err
		}
		if found {
			// Handle the collision with a warning
			log.Infof(
				`Skipping record (type/values) '%T/%v', already present in dest`,
				item,
				item,
			)
		} else {
			if !dryRun {
				destDB.Create(&item)
			}
		}
	}
	return nil
}

// findCollision will return true when the sourceItem appears to already exist in the
// destination DB
func findCollision(destDB *gorm.DB, sourceItem any) (bool, error) {
	switch typedItem := sourceItem.(type) {
	case Experiment:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"name = ?",
			typedItem.Name,
		).Count(&c)
		return c > 0, tx.Error
	case ExperimentTag:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"experiment_id = ? AND key = ?",
			typedItem.ExperimentID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Run:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ?",
			typedItem.ID,
		).Count(&c)
		return c > 0, tx.Error
	case Param:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Tag:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Metric:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case LatestMetric:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"run_uuid = ? AND key = ?",
			typedItem.RunID,
			typedItem.Key,
		).Count(&c)
		return c > 0, tx.Error
	case Dashboard:
		c := int64(0)
		tx := destDB.Model(typedItem).Where(
			"name = ? AND app_id = ?",
			typedItem.Name,
			typedItem.AppID,
		).Count(&c)
		return c > 0, tx.Error
	case App:
		c := int64(0)
		tx := destDB.Model(typedItem).Where("id = ? ", typedItem.ID).Count(&c)
		return c > 0, tx.Error
	default:
		return false, eris.New("Could not determine source item type")
	}
}
