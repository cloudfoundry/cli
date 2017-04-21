package ccerror

// InvalidRelationError is returned when an association between two entities
// cannot be created.
type InvalidRelationError struct {
	Message string
}

func (e InvalidRelationError) Error() string {
	return e.Message
}
