package translatableerror

type ApplicationUnableToStartError struct {
	AppName    string
	BinaryName string
}

func (ApplicationUnableToStartError) Error() string {
	return "Start unsuccessful\n\nTIP: use '{{.BinaryName}} logs {{.AppName}} --recent' for more information"
}

func (e ApplicationUnableToStartError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"BinaryName": e.BinaryName,
	})
}
