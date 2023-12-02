package helpers

import (
	"bytes"

	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/ipc"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func DecodeArrowMetrics(buf *bytes.Buffer) ([]models.Metric, error) {
	pool := memory.NewGoAllocator()

	// Create a new reader
	reader, err := ipc.NewReader(buf, ipc.WithAllocator(pool))
	if err != nil {
		return nil, eris.Wrap(err, "error creating reader for arrow decode")
	}
	defer reader.Release()

	var metrics []models.Metric

	// Iterate over all records in the reader
	for reader.Next() {
		rec := reader.Record()
		for i := 0; i < int(rec.NumRows()); i++ {
			metric := models.Metric{
				RunID:     rec.Column(0).(*array.String).Value(i),
				Key:       rec.Column(1).(*array.String).Value(i),
				Step:      rec.Column(2).(*array.Int64).Value(i),
				Timestamp: rec.Column(3).(*array.Int64).Value(i),
				IsNan:     rec.Column(4).(*array.Float64).IsNull(i),
			}
			if !metric.IsNan {
				metric.Value = rec.Column(4).(*array.Float64).Value(i)
			}
			metrics = append(metrics, metric)
		}
		rec.Release()
	}

	if reader.Err() != nil {
		return nil, eris.Wrap(reader.Err(), "error processing reader in arrow decode")
	}

	return metrics, nil
}
