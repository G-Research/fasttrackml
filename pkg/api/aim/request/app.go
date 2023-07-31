package request

// UpdateApp represents the data to update an App
type UpdateApp struct {
	Type  string   `json:"type"`
	State AppState `json:"state"`
}

// CreateApp represents the data to create an App
type CreateApp struct {
	Type  string   `json:"type"`
	State AppState `json:"state"`
}

// AppState represents key/value state data
type AppState map[string]any
