package helpers

import (
	"os"
	"strings"
)

const defaultDatabaseBackend = "sqlite"

func GetLogLevel() string {
	level, ok := os.LookupEnv("FML_LOG_LEVEL")
	if ok {
		return level
	}
	return "info"
}

func GetDatabaseBackend() string {
	backend, ok := os.LookupEnv("FML_DATABASE_BACKEND")
	if ok {
		return backend
	}
	dbUri, ok := os.LookupEnv("FML_DATABASE_URI")
	if ok {
		return strings.Split(dbUri, "://")[0]
	}
	return defaultDatabaseBackend
}

func GetPostgresUri() string {
	uri, ok := os.LookupEnv("FML_POSTGRES_URI")
	if ok {
		return uri
	}
	return "postgres://postgres:postgres@localhost/postgres"
}

func GetGSEndpointUri() string {
	uri, ok := os.LookupEnv("FML_GS_ENDPOINT_URI")
	if ok {
		return uri
	}
	return "http://localhost:4443/storage/v1/"
}

func GetS3EndpointUri() string {
	uri, ok := os.LookupEnv("FML_S3_ENDPOINT_URI")
	if ok {
		return uri
	}
	return "http://localhost:9000"
}

func GetSlowTestsEnabledFlag() bool {
	flag, ok := os.LookupEnv("FML_SLOW_TESTS_ENABLED")
	if ok && flag == "true" {
		return true
	}
	return false
}
