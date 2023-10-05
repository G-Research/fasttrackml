package request

import (
	"github.com/gofiber/fiber/v2"
)

// GetAlignedMetricRequest is a request object for `GET /mlflow/metrics/align` endpoint.
type GetAlignedMetricRequest struct {
	Runs    []AlignedMetricRunRequest `json:"runs"`
	AlignBy string                    `json:"align_by"`
}

// AlignedMetricRunRequest is a partial request object for GetAlignedMetricRequest
type AlignedMetricRunRequest struct {
	ID     string                      `json:"run_id"`
	Traces []AlignedMetricTraceRequest `json:"traces"`
}

// AlignedMetricTraceRequest is a partial request object for AlignedMetricRunRequest
type AlignedMetricTraceRequest struct {
	Context fiber.Map `json:"context"`
	Name    string    `json:"name"`
	Slice   []int     `json:"slice"`
}

// SearchMetricsRequest is a request struct for `GET /runs/search/metric/` endpoint.
type SearchMetricsRequest struct {
	Query          string `query:"q"`
	Steps          int    `query:"p"`
	XAxis          string `query:"x_axis"`
	SkipSystem     bool   `query:"skip_system"`
	ReportProgress bool   `query:"report_progress"`
}
