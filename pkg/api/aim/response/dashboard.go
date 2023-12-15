package response

import (
	"time"

	"github.com/google/uuid"
)

// Dashboard represents the response json in Dashboard endpoints
type Dashboard struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AppID       uuid.UUID `json:"app_id"`
	AppType     string    `json:"app_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
