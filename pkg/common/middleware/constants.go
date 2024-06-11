package middleware

import "regexp"

// regexps to detect requested API.
var (
	AdminPrefixRegexp     = regexp.MustCompile(`^/admin`)
	ChooserPrefixRegexp   = regexp.MustCompile(`^/chooser|^/$`)
	MlflowAimPrefixRegexp = regexp.MustCompile(`^/aim/api|^/ajax-api/2.0/mlflow|^/api/2.0/mlflow`)
)
