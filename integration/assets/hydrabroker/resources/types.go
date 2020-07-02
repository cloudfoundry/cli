package resources

type ServiceInstanceDetails struct {
	ServiceID  string                 `json:"service_id"`
	PlanID     string                 `json:"plan_id"`
	Parameters map[string]interface{} `json:"parameters"`
}
