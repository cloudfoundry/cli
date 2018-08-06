package translatableerror

type BuildpackNotFoundError struct {
	BuildpackName string
}

func (BuildpackNotFoundError) Error() string {
	return "Buildpack {{.BuildpackName}} not found"
}

func (e BuildpackNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName": e.BuildpackName,
	})
}
