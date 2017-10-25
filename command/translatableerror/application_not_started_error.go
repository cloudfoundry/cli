package translatableerror

type ApplicationNotStartedError struct {
	Name string
}

func (ApplicationNotStartedError) Error() string {
	return "Application '{{.AppName}}' is not in the STARTED state"
}

func (e ApplicationNotStartedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.Name,
	})
}
