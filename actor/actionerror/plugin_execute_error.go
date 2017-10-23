package actionerror

// PluginExecuteError is returned when running the plugin binary fails.
type PluginExecuteError struct {
	Err error
}

func (e PluginExecuteError) Error() string {
	return e.Err.Error()
}
