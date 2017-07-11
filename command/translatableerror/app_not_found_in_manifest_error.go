package translatableerror

type AppNotFoundInManifestError struct {
	Name string
}

func (AppNotFoundInManifestError) Error() string {
	return "Could not find app named '{{.AppName}}' in manifest"
}

func (e AppNotFoundInManifestError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.Name,
	})
}
