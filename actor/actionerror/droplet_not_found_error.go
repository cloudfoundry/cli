package actionerror

import "fmt"

// DropletNotFoundError is returned when a requested droplet from an
// application is not found.
type DropletNotFoundError struct {
	AppGUID string
}

func (e DropletNotFoundError) Error() string {
	return fmt.Sprintf("Droplet from App GUID '%s' not found.", e.AppGUID)
}
