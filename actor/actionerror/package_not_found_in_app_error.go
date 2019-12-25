package actionerror

import "fmt"

type PackageNotFoundInAppError struct {
	GUID    string
	AppName string
	BinaryName string
}

func (e PackageNotFoundInAppError) Error() string {
	switch {
	case e.GUID != "" && e.AppName != "":
		return fmt.Sprintf("Package with guid '%s' not found in app '%s'.", e.GUID, e.AppName)
	case e.AppName != "":
		return fmt.Sprintf("Package not found in app '%s'.", e.AppName)
	default:
		return fmt.Sprintf("Package not found.")
	}

}
