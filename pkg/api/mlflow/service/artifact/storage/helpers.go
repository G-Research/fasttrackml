package storage

import (
	"net/url"
	"strings"

	"github.com/rotisserie/eris"
)

// ExtractBucketAndPrefix extracts bucket name and prefix path from provided uri.
// uri could be any S3 compatible storage uri like; AWS S3 or Google Storage.
// after processing of uri s3://fasttrackml/2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts result will be:
// - bucket = fasttrackml
// - prefix = 2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts
func ExtractBucketAndPrefix(uri string) (string, string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", eris.Wrapf(err, "error parsing provided uri: %s", u)
	}

	return u.Host, strings.TrimLeft(u.Path, "/"), nil
}
