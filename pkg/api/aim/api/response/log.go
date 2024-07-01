package response

import (
	"bufio"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
)

// NewGetRunLogsResponse creates a new response object for `GET /runs/:id/logs` endpoint.
func NewGetRunLogsResponse(
	ctx *fiber.Ctx, rows *sql.Rows, next func(*sql.Rows) (*models.Log, error),
) {
	ctx.Set("Content-Type", "application/octet-stream")
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		//nolint:errcheck
		defer rows.Close()

		flush := func(w *bufio.Writer, data fiber.Map) error {
			if err := encoding.EncodeTree(w, data); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return err
			}
			return nil
		}

		start := time.Now()
		if err := func() error {
			data := fiber.Map{}
			count, batchSize := 1, 500
			for rows.Next() {
				runLog, err := next(rows)
				if err != nil {
					return eris.Wrap(err, "error getting next result")
				}
				data[fmt.Sprintf("%d", count)] = runLog.Value
				if count%batchSize == 0 {
					if err := flush(w, data); err != nil {
						return err
					}
					data = fiber.Map{}
				}
				count++
			}

			if err := flush(w, data); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			log.Errorf("error encountered in %s %s: error streaming run logs: %s", ctx.Method(), ctx.Path(), err)
		}
		log.Infof("body - %s %s %s", time.Since(start), ctx.Method(), ctx.Path())
	})
}
