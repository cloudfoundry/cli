package actionerror

// PluginBinaryRemoveFailedError is returned when running the plugin binary fails.
type PluginBinaryRemoveFailedError struct {
	Err error
}

func (e PluginBinaryRemoveFailedError) Error() string {
	return e.Err.Error()
}
