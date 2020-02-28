package actionerror

type ServicePlanVisibilityTypeError struct {
}

func (e ServicePlanVisibilityTypeError) Error() string {
	return "You cannot change access for space-scoped service plans."
}
