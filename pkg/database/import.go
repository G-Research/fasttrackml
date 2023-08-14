package database

import (
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type experimentInfo struct {
	sourceID int64
	destID   int64
}

var experimentInfos = []experimentInfo{}

// Import will copy the contents of input db to output db.
func Import(input, output *DbInstance) error {
	in := input.DB
	out := output.DB
	tables := []string{
		"experiment_tags",
		"runs",
		"tags",
		"params",
		"metrics",
		"latest_metrics",
		"apps",
		"dashboards",
	}
	// experiments needs special handling
	if err := importExperiments(in, out); err != nil {
		return eris.Wrapf(err, "error importing table %s", "experiements")
	}
	// all other tables
	for _, table := range tables {
		if err := importTable(in, out, table); err != nil {
			return eris.Wrapf(err, "error importing table %s", table)
		}
	}
	return nil
}

// importExperiments will copy the contents of one table (model) from sourceDB to destDB.
func importExperiments(sourceDB, destDB *gorm.DB) error {
	// Start transaction in the destDB
	err := destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source database
		rows, err := sourceDB.Model(Experiment{}).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var scannedItem, newItem Experiment
			if err := sourceDB.ScanRows(rows, &scannedItem); err != nil {
				return err
			}
			newItem = Experiment{
				Name:             scannedItem.Name,
				ArtifactLocation: scannedItem.ArtifactLocation,
				LifecycleStage:   scannedItem.LifecycleStage,
				CreationTime:     scannedItem.CreationTime,
				LastUpdateTime:   scannedItem.LastUpdateTime,
			}
			if err := destTX.
				Where(Experiment{Name: scannedItem.Name}).
				FirstOrCreate(&newItem).Error; err != nil {
				return err
			}
			saveExperimentInfo(scannedItem, newItem)
			count++
		}
		log.Infof("Importing %s - found %v records", "experiments", count)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// importTablewill copy the contents of one table (model) from sourceDB
// while updating the experiment_id to destDB.
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
			if err := sourceDB.ScanRows(rows, &item); err != nil {
				return err
			}
			item = translateFields(item)
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

func saveExperimentInfo(source, dest Experiment) {
	experimentInfos = append(experimentInfos, experimentInfo{
		sourceID: int64(*source.ID),
		destID:   int64(*dest.ID),
	})
}

func translateFields(item map[string]any) map[string]any {
	// boolean is numeric when coming from sqlite
	if isNaN, ok := item["is_nan"]; ok {
		switch v := isNaN.(type) {
		case bool:
			break
		default:
			item["is_nan"] = (v != 0)
		}
	}
	// items with experiment_id fk need to reference the new ID
	if expID, ok := item["experiment_id"]; ok {
		id, _ := expID.(int64)
		for _, expInfo := range experimentInfos {
			if expInfo.sourceID == id {
				item["experiment_id"] = expInfo.destID
			}
		}
	}
	return item
}
