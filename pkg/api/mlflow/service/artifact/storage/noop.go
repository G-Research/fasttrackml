package storage

// Noop represents empty storage adapter.
type Noop struct{}

// NewNoop creates new Noop instance.
func NewNoop() *Noop {
	return &Noop{}
}

// List implements Provider interface.
func (s Noop) List(_, _ string) (string, []ArtifactObject, error) {
	return "", make([]ArtifactObject, 0), nil
}
