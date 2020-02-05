package translatableerror

// QuotaNotFoundForNameError is returned when a quota with the given name can't be found.
type QuotaNotFoundForNameError struct {
	Name string
}

func (e QuotaNotFoundForNameError) Error() string {
	return "Quota {{.QuotaName}} not found"
}

func (e QuotaNotFoundForNameError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"QuotaName": e.Name,
	})
}
