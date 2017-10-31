package actionerror

// OrphanedRoutesNotFoundError is an error wrapper that represents the case
// when no orphaned routes are found.
type OrphanedRoutesNotFoundError struct{}

// Error method to display the error message.
func (OrphanedRoutesNotFoundError) Error() string {
	return "No orphaned routes were found."
}
