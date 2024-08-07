package controller

import (
	"bytes"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// convertError converts api.ErrorResponse to fiber error.
func convertError(err error) error {
	switch v := err.(type) {
	case *api.ErrorResponse:
		if v.ErrorCode == api.ErrorCodeResourceDoesNotExist {
			return fiber.ErrNotFound
		}
	}
	return err
}

func convertImagesToMap(
	images []io.ReadCloser, req request.GetRunImagesBatchRequest,
) (map[string]any, error) {
	imagesMap := make(map[string]any)

	for i, image := range images {
		var buffer bytes.Buffer
		_, err := io.CopyBuffer(&buffer, image, make([]byte, 4096))
		if err != nil {
			return nil, eris.Wrap(err, "error copying artifact Reader to output stream")
		}
		imagesMap[req[i]] = buffer.Bytes()
	}
	return imagesMap, nil
}
