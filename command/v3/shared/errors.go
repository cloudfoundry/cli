package shared

type RunTaskError struct {
	Message string
}

func (_ RunTaskError) Error() string {
	return "Error running task: {{.CloudControllerMessage}}"
}

func (e RunTaskError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CloudControllerMessage": e.Message,
	})
}

type V3APIDoesNotExistError struct {
	Message string
}

func (_ V3APIDoesNotExistError) Error() string {
	return "{{.Message}}\nNote that this command requires CF API version 3.0.0+."
}

func (e V3APIDoesNotExistError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}

type IsolationSegmentNotFoundError struct {
	Name string
}

func (_ IsolationSegmentNotFoundError) Error() string {
	return "Isolation segment '{{.Name}}' not found."
}

func (e IsolationSegmentNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}

type OrganizationNotFoundError struct {
	Name string
}

func (_ OrganizationNotFoundError) Error() string {
	return "Organization '{{.Name}}' not found."
}

func (e OrganizationNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
