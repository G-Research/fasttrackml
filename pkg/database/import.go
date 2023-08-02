package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Import will copy the contents of input db to output db.
func Import(input, output *DbInstance, dryRun bool) error {

	if err := importTable(input.DB, output.DB, dryRun, Experiment{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, ExperimentTag{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, Run{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, Tag{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, Metric{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, LatestMetric{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, Param{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, Dashboard{}); err != nil {
		return err
	}
	if err := importTable(input.DB, output.DB, dryRun, App{}); err != nil {
		return err
	}


	fmt.Println("Data transfer complete")
	return nil
}

// importTable will copy the contents of one table (model) from sourceDB to destDB
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

		// Check for a collision before attempting to insert
		// var existingRecord T
		// if err := destDB.First(&existingRecord, item.Finder()).Error; err != nil {
		// No collision, add to the slice for bulk insert
		destModels = append(destModels, item)
		// } else {
		// 	// Handle the collision with a warning (print to console in this example)
		// 	fmt.Printf("Warning: Skipping record (type/id) '%T/%s', a record with the same already exists.\n", item, )
		// }
	}

	// Perform bulk insert of non-colliding records
	if len(destModels) > 0 && !dryRun {
		destDB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(destModels, len(destModels))
	}
	return nil

}
