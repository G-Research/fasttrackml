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
