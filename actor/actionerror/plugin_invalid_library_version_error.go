package actionerror

// PluginInvalidLibraryVersionError is returned with a plugin is invalid because it is
// compiled with an incompatible version of the library.
type PluginInvalidLibraryVersionError struct {
}

func (PluginInvalidLibraryVersionError) Error() string {
	return "This plugin is not compatible with this version of the CLI"
}
