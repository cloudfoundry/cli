package translatableerror

type DropletFileError struct {
	Err error
}

func (DropletFileError) Error() string {
	return "Error creating droplet file: {{.Error}}"
}

func (e DropletFileError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}
