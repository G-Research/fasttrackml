package response

// ProjectParamsResponse is a response object for `GET aim/projects/params` endpoint.
type ProjectParamsResponse struct {
	Metric map[string][]struct{} `json:"metric"`
	Params struct {
		Tags struct {
			MlflowSourceName struct {
				ExampleType string `json:"__example_type__"`
			} `json:"mlflow.source.name"`
			MlflowSourceType struct {
				ExampleType string `json:"__example_type__"`
			} `json:"mlflow.source.type"`
			MlflowUser struct {
				ExampleType string `json:"__example_type__"`
			} `json:"mlflow.user"`
		} `json:"tags"`
	} `json:"params"`
}
