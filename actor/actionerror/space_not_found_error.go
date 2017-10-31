package actionerror

import "fmt"

// SpaceNotFoundError represents the scenario when the space searched for could
// not be found.
type SpaceNotFoundError struct {
	GUID string
	Name string
}

func (e SpaceNotFoundError) Error() string {
	switch {
	case e.Name != "":
		return fmt.Sprintf("Space '%s' not found.", e.Name)
	case e.GUID != "":
		return fmt.Sprintf("Space with GUID '%s' not found.", e.GUID)
	default:
		return fmt.Sprintf("Space '' not found.")
	}
}
