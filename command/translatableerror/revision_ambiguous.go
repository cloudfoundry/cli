package translatableerror

// RevisionAmbiguousError is returned when multiple revisions with the same
// version are returned
type RevisionAmbiguousError struct {
	Version int
}

func (e RevisionAmbiguousError) Error() string {
	return "More than one revision '{{.RevisionVersion}}' found"
}

func (e RevisionAmbiguousError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RevisionVersion": e.Version,
	})
}
