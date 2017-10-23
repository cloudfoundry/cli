package actionerror

// AssignDropletError is returned when assigning the current droplet of an app
// fails
type AssignDropletError struct {
	Message string
}

func (a AssignDropletError) Error() string {
	return a.Message
}
