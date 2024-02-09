package repositories

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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
	//nolint:errcheck
	defer mockDb.Close()

	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.Nil(t, err)

	repo := NewRunRepository(db)

	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			if tc.startWith < 0 {
				assert.EqualError(
					t,
					repo.renumberRows(db, tc.startWith),
					"attempting to renumber with less than 0 row number value",
				)
			} else {
				mock.ExpectExec(
					"LOCK TABLE runs",
				).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`UPDATE runs`).WithArgs(
					tc.startWith, tc.startWith,
				).WillReturnResult(
					sqlmock.NewResult(0, 1),
				)
				assert.NoError(t, repo.renumberRows(db, tc.startWith))
			}
		})
	}
}
