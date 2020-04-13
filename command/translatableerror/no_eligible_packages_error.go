package translatableerror

type NoEligiblePackagesError struct {
	AppName    string
	BinaryName string
}

func (NoEligiblePackagesError) Error() string {
	return "App '{{.AppName}}' has no eligible packages.\n\nTIP: Use '{{.BinaryName}} packages {{.AppName}}' to list packages in your app. Use '{{.BinaryName}} create-package' to create one."
}

func (e NoEligiblePackagesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"BinaryName": e.BinaryName,
	})
}
