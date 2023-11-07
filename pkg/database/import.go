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
	// items with experiment_id fk need to reference the new ID
	if expID, ok := item["experiment_id"]; ok {
		id, ok := expID.(int64)
		if !ok {
			return nil, eris.Errorf("unable to assert experiment_id as int64: %d", expID)
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
		if srcID, ok := item[field]; ok {
			stringID := fmt.Sprintf("%v", reflect.Indirect(reflect.ValueOf(srcID)))
			if uuidRegexp.MatchString(stringID) {
				binID, err := uuid.Parse(stringID)
				if err != nil {
					return nil, eris.Errorf("unable to create binary UUID field from string: %s", stringID)
				}
				item[field] = binID
			}
		}
	}
	return item, nil
}
