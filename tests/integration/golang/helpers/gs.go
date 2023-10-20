package helpers

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/api/googleapi"

	"cloud.google.com/go/storage"
	"github.com/rotisserie/eris"
	"google.golang.org/api/iterator"
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

// PrepareTestBuckets prepares test buckets for further usage in tests:
// - delete provided buckets and all the content within them.
// - create the same empty bucket.
func PrepareTestBuckets(client *storage.Client, buckets []string) error {
	for _, bucket := range buckets {
		it := client.Bucket(bucket).Objects(context.Background(), &storage.Query{})
		for {
			object, err := it.Next()
			if errors.Is(err, iterator.Done) || errors.Is(err, storage.ErrBucketNotExist) {
				break
			}
			if err := client.Bucket(bucket).Object(object.Name).Delete(context.Background()); err != nil {
				return err
			}
		}
		var e *googleapi.Error

		if err := client.Bucket(bucket).Delete(context.Background()); errors.As(err, &e) &&
			e.Code != http.StatusNotFound {
			return err
		}
		if err := client.Bucket(bucket).Create(context.Background(), "", nil); err != nil {
			return err
		}
	}
	return nil
}
