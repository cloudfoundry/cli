package translatableerror

type PackageNotFoundInAppError struct {
	AppName    string
	BinaryName string
}

func (PackageNotFoundInAppError) Error() string {
	return "Package not found in app '{{.AppName}}'.\n\nTIP: Use '{{.BinaryName}} packages {{.AppName}}' to list packages in your app. Use '{{.BinaryName}} create-package' to create one."
}

func (e PackageNotFoundInAppError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"BinaryName": e.BinaryName,
	})
}
