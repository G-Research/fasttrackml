package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"	
	"gopkg.in/DATA-DOG/go-sqlmock.v1"	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestRenumberRowsForNegativeRowNumber(t *testing.T) {
	mockDb, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDb.Close()
	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	   })
	db, _ := gorm.Open(dialector, &gorm.Config{})

	startWith := models.RowNum(-1)

	repo := NewRunRepository(db)
	err = repo.renumberRows(db, startWith)

	assert.EqualError(t, err, "attempting to renumber with less than 0 row number value")
}