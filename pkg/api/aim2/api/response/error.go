package response

// Error represents the response json in api errors
type Error struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
}
