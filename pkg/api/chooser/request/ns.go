package request

// CreateNamespace represents the data to create an Namespace.
type CreateNamespace struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
