package ccerror

// SpaceNameTakenError is returned when creating a
// space that already exists.
type SpaceNameTakenError struct {
	Message string
}

func (e SpaceNameTakenError) Error() string {
	return e.Message
}
