package translatableerror

// NoDropletForAppError will default to current droplet if no droplet GUID is provided
type NoDropletForAppError struct {
	AppName     string
	DropletGUID string
}

func (e NoDropletForAppError) Error() string {
	if e.DropletGUID != "" {
		return "Droplet '{{.DropletGUID}}' not found for app '{{.AppName}}'"
	}
	return "App '{{.AppName}}' does not have a current droplet."
}

func (e NoDropletForAppError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":     e.AppName,
		"DropletGUID": e.DropletGUID,
	})
}
