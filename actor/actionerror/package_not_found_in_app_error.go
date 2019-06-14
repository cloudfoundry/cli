package actionerror

import "fmt"

type PackageNotFoundInAppError struct {
	GUID    string
	AppName string
}

func (e PackageNotFoundInAppError) Error() string {
	return fmt.Sprintf("Package with guid '%s' not found in app '%s'.", e.GUID, e.AppName)
}
