package translatableerror

type NoCompatibleBinaryError struct {
}

func (e NoCompatibleBinaryError) Error() string {
	return "Plugin requested has no binary available for your platform."
}

func (e NoCompatibleBinaryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
