package repositories

import (
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Test_renumberRows(t *testing.T) {
	testData := []struct {
		name      string
		startWith models.RowNum
	}{
		{
			name:      "NegativeRowNumber",
			startWith: models.RowNum(-1),
		},
		{
			name:      "ZeroRowNumber",
			startWith: models.RowNum(0),
		},
		{
			name:      "PositiveRowNumber",
			startWith: models.RowNum(1),
		},
	}

	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDb.Close()

	lockExpect := func() {
		mock.ExpectExec("LOCK TABLE runs").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE runs`).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.Nil(t, err)

	repo := NewRunRepository(db)

	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			if tc.startWith < 0 {
				err := repo.renumberRows(db, tc.startWith)
				assert.EqualError(t, err, "attempting to renumber with less than 0 row number value")
			} else {
				lockExpect()
				err := repo.renumberRows(db, tc.startWith)
				assert.NoError(t, err)
			}
		})
	}
}
