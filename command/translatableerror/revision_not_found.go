package translatableerror

// RevisionNotFoundError is returned when a requested revision is not
// found.
type RevisionNotFoundError struct {
	Version int
}

func (e RevisionNotFoundError) Error() string {
	return "Revision '{{.RevisionVersion}}' not found"
}

func (e RevisionNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RevisionVersion": e.Version,
	})
}
