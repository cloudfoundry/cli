package translatableerror

type InvalidYAMLError struct {
	Err error
}

func (e InvalidYAMLError) Error() string {
	return "The option --vars-file expects a valid YAML file. {{.ErrorMessage}}"
}

func (e InvalidYAMLError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ErrorMessage": e.Err.Error(),
	})
}
