package helpers

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/rotisserie/eris"
	"google.golang.org/api/option"
)

// NewGSClient creates new instance of Google Storage client.
func NewGSClient(endpoint string) (*storage.Client, error) {
	client, err := storage.NewClient(
		context.TODO(), option.WithEndpoint(endpoint), option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, eris.Wrap(err, "error creating GS client")
	}

	return client, nil
}
