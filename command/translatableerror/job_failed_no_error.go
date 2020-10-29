package translatableerror

// RevisionAmbiguousError is returned when multiple revisions with the same
// version are returned
type JobFailedNoErrorError struct {
	JobGUID string
}

func (e JobFailedNoErrorError) Error() string {
	return "Job {{.JobGUID}} failed with no error. This is unexpected, contact your operator for details."
}

func (e JobFailedNoErrorError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"JobGUID": e.JobGUID,
	})
}
