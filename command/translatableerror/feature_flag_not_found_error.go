package translatableerror

type FeatureFlagNotFoundError struct {
	Name string
}

func (e FeatureFlagNotFoundError) Error() string {
	return "Feature Flag {{.FlagName}} not found"
}

func (e FeatureFlagNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"FlagName": e.Name,
	})
}
