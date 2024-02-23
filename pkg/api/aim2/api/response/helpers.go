package response

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/gofiber/fiber/v2"
)

func toNumpy(values []float64) fiber.Map {
	buf := bytes.NewBuffer(make([]byte, 0, len(values)*8))
	for _, v := range values {
		switch v {
		case math.MaxFloat64:
			v = math.Inf(1)
		case -math.MaxFloat64:
			v = math.Inf(-1)
		}
		//nolint:gosec,errcheck
		binary.Write(buf, binary.LittleEndian, v)
	}
	return fiber.Map{
		"type":  "numpy",
		"dtype": "float64",
		"shape": len(values),
		"blob":  buf.Bytes(),
	}
}
