package translatableerror

// ProcessInstanceNotFoundError is returned when a proccess type or process instance can't be found
type ProcessInstanceNotFoundError struct {
	ProcessType   string
	InstanceIndex uint
}

func (ProcessInstanceNotFoundError) Error() string {
	return "Instance {{.InstanceIndex}} of process {{.ProcessType}} not found"
}

func (e ProcessInstanceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ProcessType":   e.ProcessType,
		"InstanceIndex": e.InstanceIndex,
	})
}
