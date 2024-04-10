package oidc

import "github.com/G-Research/fasttrackml/pkg/common/config/auth"

// Client represents OIDC client.
type Client struct{}

// NewClient creates new OIDC client
func NewClient(config *auth.Config) *Client {
	return &Client{}
}
