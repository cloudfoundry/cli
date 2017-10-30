package translatableerror

// ProcessInstanceNotRunningError is returned when trying to perform an action
// on an instance that is not running
type ProcessInstanceNotRunningError struct {
	ProcessType   string
	InstanceIndex uint
}

func (ProcessInstanceNotRunningError) Error() string {
	return "Instance {{.InstanceIndex}} of process {{.ProcessType}} not running"
}

func (e ProcessInstanceNotRunningError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ProcessType":   e.ProcessType,
		"InstanceIndex": e.InstanceIndex,
	})
}
