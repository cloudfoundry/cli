package ccerror

// RoleAlreadyExistsError is returned when a role with the same type, user,
// and org or space already exists in the Cloud Controller.
type RoleAlreadyExistsError struct {
	UnprocessableEntityError
}
