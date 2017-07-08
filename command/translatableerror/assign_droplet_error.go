package translatableerror

// AssignDropletError is returned when assigning the current droplet of an app
// fails
type AssignDropletError struct {
	Message string
}

func (AssignDropletError) Error() string {
	return "Unable to assign droplet: {{.CloudControllerMessage}}"
}

func (e AssignDropletError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CloudControllerMessage": e.Message,
	})
}
