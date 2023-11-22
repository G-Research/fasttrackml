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

// CreateGSBuckets creates the tests buckets.
func CreateGSBuckets(client *storage.Client, buckets []string) error {
	for _, bucket := range buckets {
		if err := client.Bucket(bucket).Create(context.Background(), "", nil); err != nil {
			return eris.Wrapf(err, "failed to create bucket %q", bucket)
		}
	}
	return nil
}

// DeleteGSBuckets deletes the tests buckets.
func DeleteGSBuckets(client *storage.Client, buckets []string) error {
	for _, bucket := range buckets {
		it := client.Bucket(bucket).Objects(context.Background(), &storage.Query{})
		for {
			object, err := it.Next()
			if err != nil {
				if errors.Is(err, iterator.Done) || errors.Is(err, storage.ErrBucketNotExist) {
					break
				}
				return eris.Wrapf(err, "failed to list objects in bucket %q", bucket)
			}
			if err := client.Bucket(bucket).Object(object.Name).Delete(context.Background()); err != nil {
				return eris.Wrapf(err, "failed to delete objects in bucket %q", bucket)
			}
		}
		var e *googleapi.Error
		if err := client.Bucket(bucket).Delete(context.Background()); errors.As(err, &e) &&
			e.Code != http.StatusNotFound {
			return eris.Wrapf(err, "failed to delete bucket %q", bucket)
		}
	}
	return nil
}
