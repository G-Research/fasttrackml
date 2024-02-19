package dto

// MetricKeysItemDTO represents DTO object to temporary store data.
type MetricKeysItemDTO struct {
	Name    string
	Context string
}

// MetricKeysMapDTO represents DTO object to temporary store data.
type MetricKeysMapDTO map[MetricKeysItemDTO]any
