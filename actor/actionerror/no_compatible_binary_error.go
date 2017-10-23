package actionerror

// NoCompatibleBinaryError is returned when a repository contains a specified
// plugin but not for the specified platform.
type NoCompatibleBinaryError struct {
}

func (e NoCompatibleBinaryError) Error() string {
	return "Plugin requested has no binary available for your platform."
}
