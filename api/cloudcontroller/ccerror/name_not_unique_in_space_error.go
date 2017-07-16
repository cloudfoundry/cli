package ccerror

type NameNotUniqueInSpaceError struct {
}

func (e NameNotUniqueInSpaceError) Error() string {
	return "name must be unique in space"
}
