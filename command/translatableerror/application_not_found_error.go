package translatableerror

type ApplicationNotFoundError struct {
	GUID string
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	if e.GUID != "" {
		return "Application with GUID {{.GUID}} not found."
	}
	return "App {{.AppName}} not found"
}

func (e ApplicationNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"GUID":    e.GUID,
		"AppName": e.Name,
	})
}
