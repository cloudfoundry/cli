package translatableerror

type ApplicationNotFoundError struct {
	Name string
}

func (_ ApplicationNotFoundError) Error() string {
	return "App {{.AppName}} not found"
}

func (e ApplicationNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.Name,
	})
}
