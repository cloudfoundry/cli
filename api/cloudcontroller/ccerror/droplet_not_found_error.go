package ccerror

// DropletNotFoundError is returned when an endpoint cannot find the
// specified application
type DropletNotFoundError struct {
}

func (e DropletNotFoundError) Error() string {
	return "Droplet not found"
}
