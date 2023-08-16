package request

type ViewType string

const (
	ViewTypeAll         ViewType = "ALL"
	ViewTypeActiveOnly  ViewType = "ACTIVE_ONLY"
	ViewTypeDeletedOnly ViewType = "DELETED_ONLY"
)

type PageToken struct {
	Offset int32 `json:"offset"`
}
