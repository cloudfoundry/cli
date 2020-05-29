package translatableerror

// NoCurrentDropletForAppError is returned when there is no current droplet for an app
type NoCurrentDropletForAppError struct {
	AppName string
}

func (NoCurrentDropletForAppError) Error() string {
	return "App '{{.AppName}}' does not have a current droplet."
}

func (e NoCurrentDropletForAppError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.AppName,
	})
}
