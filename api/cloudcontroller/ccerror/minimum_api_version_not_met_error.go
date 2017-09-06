package ccerror

import "fmt"

type MinimumAPIVersionNotMetError struct {
	CurrentVersion string
	MinimumVersion string
}

func (e MinimumAPIVersionNotMetError) Error() string {
	return fmt.Sprintf("CF API version %s or higher is required. Your target is %s.", e.CurrentVersion, e.MinimumVersion)
}
