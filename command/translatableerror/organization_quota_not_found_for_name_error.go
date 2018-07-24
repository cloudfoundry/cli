package translatableerror

// OrganizationQuotaNotFoundForNameError is returned when a quota with the given name can't be found.
type OrganizationQuotaNotFoundForNameError struct {
	Name string
}

func (e OrganizationQuotaNotFoundForNameError) Error() string {
	return "Quota {{.QuotaName}} not found"
}

func (e OrganizationQuotaNotFoundForNameError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"QuotaName": e.Name,
	})
}
