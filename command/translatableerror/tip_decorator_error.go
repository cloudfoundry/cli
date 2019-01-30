package translatableerror

type TipDecoratorError struct {
	Tip       string
	TipKeys   map[string]interface{}
	BaseError error
}

func (e TipDecoratorError) Error() string {
	return "{{.BaseError}}\n\nTIP: {{.Tip}}"
}

func (e TipDecoratorError) Translate(translate func(string, ...interface{}) string) string {
	baseError := e.BaseError.Error()
	if translatableBaseError, ok := e.BaseError.(TranslatableError); ok {
		baseError = translatableBaseError.Translate(translate)
	}

	tip := translate(e.Tip, e.TipKeys)

	return translate(e.Error(), map[string]interface{}{
		"BaseError": baseError,
		"Tip":       tip,
	})
}
