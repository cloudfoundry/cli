package ccerror

type NameNotUniqueInOrgError struct {
}

func (e NameNotUniqueInOrgError) Error() string {
	return "name must be unique per organization"
}
