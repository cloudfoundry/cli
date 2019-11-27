package ccerror

// RoleNotFoundError is returned when a role does not exist.
type RoleNotFoundError struct {
}

func (e RoleNotFoundError) Error() string {
	return "Role not found"
}
