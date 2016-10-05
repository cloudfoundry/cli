package cloudcontrollerv2

type PaginatedWrapper struct {
	NextURL   string      `json:"next_url"`
	Resources interface{} `json:"resources"`
}
