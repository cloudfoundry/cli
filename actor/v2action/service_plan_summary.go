package v2action

type ServicePlanSummary struct {
	ServicePlan
	// VisibleTo is a list of Organization names that have access to this service
	// plan.
	VisibleTo []string
}
