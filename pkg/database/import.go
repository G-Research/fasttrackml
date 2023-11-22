package database

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
)

type experimentInfo struct {
	destID   int64
	sourceID int64
}

var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Importer will handle transport of data from source to destination db.
type Importer struct {
	destDB          *gorm.DB
	sourceDB        *gorm.DB
	experimentInfos []experimentInfo
}

// NewImporter initializes an Importer.
func NewImporter(input, output *gorm.DB) *Importer {
	return &Importer{
		destDB:          output,
		sourceDB:        input,
		experimentInfos: []experimentInfo{},
	}
}

// Import copies the contents of input db to output db.
func (s *Importer) Import() error {
	tables := []string{
		"namespaces",
		"apps",
		"dashboards",
		"experiments",
		"experiment_tags",
		"runs",
		"tags",
		"params",
		"metrics",
		"latest_metrics",
	}
	for _, table := range tables {
		if err := s.importTable(table); err != nil {
			return eris.Wrapf(err, "error importing table %s", table)
		}
	}
	if err := s.updateNamespaceDefaultExperiment(); err != nil {
		return eris.Wrap(err, "error updating namespace default experiment")
	}
	return nil
}

// importExperiments copies the contents of the experiments table from sourceDB to destDB,
// while recording the new ID.
func (s *Importer) importExperiments() error {
	// Start transaction in the destDB
	err := s.destDB.Transaction(func(destTX *gorm.DB) error {
		// Query data from the source database
		rows, err := s.sourceDB.Model(Experiment{}).Rows()
		if err != nil {
			return eris.Wrap(err, "error creating Rows instance from source")
		}
		if err := rows.Err(); err != nil {
			return eris.Wrap(err, "error getting query result")
		}
		//nolint:errcheck
		defer rows.Close()

		count := 0
		for rows.Next() {
			var scannedItem Experiment
			if err := s.sourceDB.ScanRows(rows, &scannedItem); err != nil {
				return eris.Wrap(err, "error creating Rows instance from source")
			}
			newItem := Experiment{
				Name:             scannedItem.Name,
				NamespaceID:      scannedItem.NamespaceID,
				ArtifactLocation: scannedItem.ArtifactLocation,
				LifecycleStage:   scannedItem.LifecycleStage,
				CreationTime:     scannedItem.CreationTime,
				LastUpdateTime:   scannedItem.LastUpdateTime,
			}
			// keep default experiment ID, but otherwise draw new one
			if *scannedItem.ID == int32(0) {
				newItem.ID = scannedItem.ID
			}
			if err := destTX.
				Where(Experiment{Name: scannedItem.Name}).
				FirstOrCreate(&newItem).Error; err != nil {
				return eris.Wrap(err, "error creating destination row")
			}
			s.saveExperimentInfo(scannedItem, newItem)
			count++
		}
		log.Infof("Importing experiments - found %d records", count)
		return nil
	})
	if err != nil {
		return eris.Wrap(err, "error copying experiments table")
	}
	return nil
}

// importTable copies the contents of one table (model) from sourceDB
// while updating the experiment_id to destDB.
func (s *Importer) importTable(table string) error {
	switch table {
	// handle special case for experiments.
	case "experiments":
		if err := s.importExperiments(); err != nil {
			return eris.Wrap(err, "error importing table experiments")
		}
	default:
		// Start transaction in the destDB
		err := s.destDB.Transaction(func(destTX *gorm.DB) error {
			// Query data from the source database
			rows, err := s.sourceDB.Table(table).Rows()
			if err != nil {
				return eris.Wrap(err, "error creating Rows instance from source")
			}
			if err := rows.Err(); err != nil {
				return eris.Wrap(err, "error getting query result")
			}
			//nolint:errcheck
			defer rows.Close()

			count := 0
			for rows.Next() {
				var item map[string]any
				if err = s.sourceDB.ScanRows(rows, &item); err != nil {
					return eris.Wrap(err, "error scanning source row")
				}
				item, err = s.translateFields(item)
				if err != nil {
					return eris.Wrap(err, "error translating fields")
				}
				if err := destTX.
					Table(table).
					Clauses(clause.OnConflict{DoNothing: true}).
					Create(&item).Error; err != nil {
					return eris.Wrap(err, "error creating destination row")
				}
				count++
			}
			log.Infof("Importing %s - found %d records", table, count)
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// saveExperimentInfo maps source and destination experiment for later id mapping.
func (s *Importer) saveExperimentInfo(source, dest Experiment) {
	s.experimentInfos = append(s.experimentInfos, experimentInfo{
		destID:   int64(*dest.ID),
		sourceID: int64(*source.ID),
	})
}

// translateFields alters row before creation as needed (especially, replacing old experiment_id with new).
func (s *Importer) translateFields(item map[string]any) (map[string]any, error) {
	// boolean fields are numeric when coming from sqlite
	booleanFields := []string{"is_nan", "is_archived"}
	for _, field := range booleanFields {
		if fieldVal, ok := item[field]; ok {
			switch v := fieldVal.(type) {
			case bool:
				break
			default:
				item[field] = v != 0.0
			}
		}
	}
	// items with experiment_id need to reference the new ID
	if expID, ok := item["experiment_id"]; ok {
		id, ok := expID.(int64)
		if !ok {
			return nil, eris.Errorf("unable to assert %s as int64: %d", "experiment_id", expID)
		}
		for _, expInfo := range s.experimentInfos {
			if expInfo.sourceID == id {
				item["experiment_id"] = expInfo.destID
			}
		}
	}
	// items with string uuid need to translate to UUID native type
	uuidFields := []string{"id", "app_id"}
	for _, field := range uuidFields {
		if srcUUID, ok := item[field]; ok {
			// when uuid, this field will be pointer to interface{} and requires some reflection
			stringUUID := fmt.Sprintf("%v", reflect.Indirect(reflect.ValueOf(srcUUID)))
			if uuidRegexp.MatchString(stringUUID) {
				binID, err := uuid.Parse(stringUUID)
				if err != nil {
					return nil, eris.Errorf("unable to create binary UUID field from string: %s", stringUUID)
				}
				item[field] = binID
			}
		}
	}
	return item, nil
}

// updateNamespaceDefaultExperiment updates the default_experiment_id for all namespaces
// when its related experiment received a new id.
func (s Importer) updateNamespaceDefaultExperiment() error {
	// Start transaction in the destDB
	err := s.destDB.Transaction(func(destTX *gorm.DB) error {
		// Get namespaces
		var namespaces []Namespace
		if err := destTX.Model(Namespace{}).Find(&namespaces).Error; err != nil {
			return eris.Wrap(err, "error reading namespaces in destination")
		}
		for _, ns := range namespaces {
			updatedExperimentID := ns.DefaultExperimentID
			for _, expInfo := range s.experimentInfos {
				if ns.DefaultExperimentID != nil && expInfo.sourceID == int64(*ns.DefaultExperimentID) {
					updatedExperimentID = common.GetPointer[int32](int32(expInfo.destID))
					break
				}
			}
			if err := destTX.
				Model(Namespace{}).
				Where(Namespace{ID: ns.ID}).
				Update("default_experiment_id", updatedExperimentID).Error; err != nil {
				return eris.Wrap(err, "error updating destination namespace row")
			}
		}
		log.Infof("Updating namespaces - processed %d records", len(namespaces))
		return nil
	})
	return err
}
