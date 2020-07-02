package translatableerror

type CurlExit22Error struct {
	StatusCode int
}

func (e CurlExit22Error) Error() string {
	return "The requested URL returned error: {{.StatusCode}}"
}

func (e CurlExit22Error) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(),
		map[string]interface{}{"StatusCode": e.StatusCode},
	)
}
