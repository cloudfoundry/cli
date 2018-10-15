package ccv2

// ServiceExtra contains extra service related properties.
type ServiceExtra struct {
	// Shareable is true if the service is shareable across organizations and
	// spaces.
	Shareable bool

	DocumentationURL string `json:"documentationUrl"`
}
