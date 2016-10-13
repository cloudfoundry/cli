package ccv2

// PaginatedWrapper represents the standard pagination format of a request that
// returns back more than one object.
type PaginatedWrapper struct {
	NextURL   string      `json:"next_url"`
	Resources interface{} `json:"resources"`
}
