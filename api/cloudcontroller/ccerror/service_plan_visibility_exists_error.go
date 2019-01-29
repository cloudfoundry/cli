package ccerror

// ServicePlanVisibilityExistsError is returned when creating a
// service plan visibility that already exists
type ServicePlanVisibilityExistsError struct {
	Message string
}

func (e ServicePlanVisibilityExistsError) Error() string {
	return e.Message
}
