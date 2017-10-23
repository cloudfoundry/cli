package actionerror

// PluginInvalidError is returned with a plugin is invalid because it is
// missing a name or has 0 commands.
type PluginInvalidError struct {
	Err error
}

func (PluginInvalidError) Error() string {
	return "File is not a valid cf CLI plugin binary."
}
