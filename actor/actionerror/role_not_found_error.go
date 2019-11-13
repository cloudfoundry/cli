package actionerror

// RoleNotFoundError is returned when a matching role is not found
type RoleNotFoundError struct {
}

func (e RoleNotFoundError) Error() string {
	return "Role not found"
}
