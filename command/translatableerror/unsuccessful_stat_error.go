package translatableerror

type UnsuccessfulStartError struct {
	AppName    string
	BinaryName string
}

func (UnsuccessfulStartError) Error() string {
	return "Start unsuccessful\n\nTIP: use '{{.BinaryName}} logs {{.AppName}} --recent' for more information"
}

func (e UnsuccessfulStartError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"BinaryName": e.BinaryName,
	})
}
