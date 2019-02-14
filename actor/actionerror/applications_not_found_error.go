package actionerror

// ApplicationsNotFoundError is returned when requested applications are not
// found.
type ApplicationsNotFoundError struct {
}

func (e ApplicationsNotFoundError) Error() string {
	return "One or more applications were not found."
}
