package actionerror

// UserNotFoundError is an error wrapper that represents the case
// when the user is not found in UAA.
type UAAUserNotFoundError struct {
}

// Error method to display the error message.
func (e UAAUserNotFoundError) Error() string {
	return "User not found."
}
