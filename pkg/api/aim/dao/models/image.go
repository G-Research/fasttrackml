package models

type ImageValues struct {
	BlobURI string      `json:"blob_uri"`
	Caption string      `json:"caption"`
	Context interface{} `json:"context"`
	Format  string      `json:"format"`
	Height  interface{} `json:"height"`
	Index   interface{} `json:"index"`
	Key     string      `json:"key"`
	SeqKey  string      `json:"seqKey"`
	Name    string      `json:"name"`
	Run     interface{} `json:"run"`
	Step    int         `json:"step"`
	Width   interface{} `json:"width"`
}

type Image struct {
	RecordRange interface{}     `json:"record_range"`
	IndexRange  interface{}     `json:"index_range"`
	Name        string          `json:"name"`
	Context     interface{}     `json:"context"`
	Values      [][]ImageValues `json:"values"`
	Iters       int             `json:"iters"`
}
