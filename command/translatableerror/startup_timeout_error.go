package translatableerror

type StartupTimeoutError struct {
	AppName    string
	BinaryName string
}

func (StartupTimeoutError) Error() string {
	return "Start app timeout\n\nTIP: Application must be listening on the right port. Instead of hard coding the port, use the $PORT environment variable.\n\nUse '{{.BinaryName}} logs {{.AppName}} --recent' for more information"
}

func (e StartupTimeoutError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"BinaryName": e.BinaryName,
	})
}
