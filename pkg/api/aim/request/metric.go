package request

import (
	"github.com/gofiber/fiber/v2"
)

// GetAlignedMetricRequest is a request object for `GET /mlflow/metrics/align` endpoint.
type GetAlignedMetricRequest struct {
	AlignBy string                    `json:"align_by"`
	Runs    []AlignedMetricRunRequest `json:"runs"`
}

type AlignedMetricRunRequest struct {
	ID     string                      `json:"run_id"`
	Traces []AlignedMetricTraceRequest `json:"traces"`
}

type AlignedMetricTraceRequest struct {
	Context fiber.Map `json:"context"`
	Name    string    `json:"name"`
	Slice   [3]int    `json:"slice"`
}
