package storage

// Nope represents empty storage adapter.
type Nope struct{}

// NewNope creates new Nope instance.
func NewNope() *Nope {
	return &Nope{}
}

// List implements Provider interface.
func (s Nope) List(path, nextPageToken string) (string, string, []ArtifactObject, error) {
	return "", "", make([]ArtifactObject, 0), nil
}
