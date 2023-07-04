package storage

import "strings"

// PrepareS3Prefix transforms input path to correct S3 prefix.
// `runArtifactURI` represents path in next format -> s3://fasttrackml/2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts.
// to correctly pass it to S3, we have to remove `s3://fasttrackml/` prefix,
// so final result has to be -> 2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts.
func PrepareS3Prefix(basePath, runArtifactURI string) string {
	return strings.TrimLeft(strings.Replace(runArtifactURI, basePath, "", -1), "/")
}
