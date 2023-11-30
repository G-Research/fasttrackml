package helpers

import "os"

func GetLogLevel() string {
	level, ok := os.LookupEnv("FML_LOG_LEVEL")
	if ok {
		return level
	}
	return "info"
}

func GetDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite:///tmp/fasttrackml.db"
}

func GetServiceUri() string {
	uri, ok := os.LookupEnv("FML_SERVICE_URI")
	if ok {
		return uri
	}
	return "http://localhost:5000"
}

func GetS3EndpointUri() string {
	uri, ok := os.LookupEnv("FML_S3_ENDPOINT_URI")
	if ok {
		return uri
	}
	return "http://localhost:9000"
}

func GetGSEndpointUri() string {
	uri, ok := os.LookupEnv("FML_GS_ENDPOINT_URI")
	if ok {
		return uri
	}
	return "http://localhost:4443/storage/v1/"
}

func GetInputDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_INPUT_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite:///tmp/fasttrackml-in.db"
}

func GetOutputDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_OUTPUT_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite:///tmp/fasttrackml-out.db"
}
