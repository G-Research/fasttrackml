package helpers

import "os"

func GetDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite://fasttrackml.db"
}

func GetServiceUri() string {
	uri, ok := os.LookupEnv("FML_SERVICE_URI")
	if ok {
		return uri
	}
	return "http://localhost:5000"
}

func GetInputDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_INPUT_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite://fasttrackml-in.db"
}

func GetOutputDatabaseUri() string {
	uri, ok := os.LookupEnv("FML_OUTPUT_DATABASE_URI")
	if ok {
		return uri
	}
	return "sqlite://fasttrackml-out.db"
}
