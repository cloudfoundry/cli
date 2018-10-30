package ccv2

// ServiceExtra contains extra service related properties.
type ServiceExtra struct {
	// DocumentationURL is the URL of the documentation page for the service.
	DocumentationURL string `json:"documentationUrl"`

	// Shareable is true if the service is shareable across organizations and
	// spaces.
	Shareable bool
}
