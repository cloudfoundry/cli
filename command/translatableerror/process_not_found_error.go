package translatableerror

// ProcessNotFoundError is returned when a proccess type can't be found
type ProcessNotFoundError struct {
	ProcessType string
}

func (ProcessNotFoundError) Error() string {
	return "Process {{.ProcessType}} not found"
}

func (e ProcessNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ProcessType": e.ProcessType,
	})
}
