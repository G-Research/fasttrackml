package request

// SearchMetricsRequest is a request struct for `GET /runs/search/metric/` endpoint.
type SearchMetricsRequest struct {
	Query          string `query:"q"`
	Steps          int    `query:"p"`
	XAxis          string `query:"x_axis"`
	SkipSystem     bool   `query:"skip_system"`
	ReportProgress bool   `query:"report_progress"`
}
