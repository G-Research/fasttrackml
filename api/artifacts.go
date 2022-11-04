package api

import (
	"fasttrack/model"
	"net/http"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func ArtifactList(db *gorm.DB) HandlerFunc {
	return EnsureMethod(func(w http.ResponseWriter, r *http.Request) any {
		id := r.URL.Query().Get("run_id")
		if id == "" {
			id = r.URL.Query().Get("run_uuid")
		}
		path := r.URL.Query().Get("path")
		token := r.URL.Query().Get("page_token")

		log.Debugf("ArtifactList request: run_id='%s', path='%s', page_token='%s'", id, path, token)

		if id == "" {
			return NewError(ErrorCodeInvalidParameterValue, "Missing value for required parameter 'run_id'")
		}

		run := model.Run{
			ID: id,
		}

		if tx := db.Select("artifact_uri").First(&run); tx.Error != nil {
			return NewError(ErrorCodeInternalError, "Unable to get artifact URI for run '%s'", id)
		}

		// TODO grab list of artifacts from S3
		resp := &ArtifactListResponse{}

		log.Debugf("ArtifactList response: %#v", resp)

		return resp
	},
		http.MethodGet,
	)
}
