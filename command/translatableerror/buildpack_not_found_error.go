package translatableerror

type BuildpackNotFoundError struct {
	BuildpackName string
	StackName     string
}

func (e BuildpackNotFoundError) Error() string {
	if len(e.StackName) == 0 {
		return "Buildpack {{.BuildpackName}} not found"
	}
	return "Buildpack {{.BuildpackName}} with stack {{.StackName}} not found"
}

func (e BuildpackNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName": e.BuildpackName,
		"StackName":     e.StackName,
	})
}
