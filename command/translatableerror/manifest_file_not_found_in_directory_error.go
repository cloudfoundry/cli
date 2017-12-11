package translatableerror

type ManifestFileNotFoundInDirectoryError struct {
	PathToManifest string
}

func (ManifestFileNotFoundInDirectoryError) Error() string {
	return "Could not find 'manifest.yml' file in {{.PathToManifest}}"
}

func (e ManifestFileNotFoundInDirectoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PathToManifest": e.PathToManifest,
	})
}
