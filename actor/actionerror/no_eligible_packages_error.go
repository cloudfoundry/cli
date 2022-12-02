package actionerror

import "fmt"

type NoEligiblePackagesError struct {
	AppName    string
	BinaryName string
}

func (e NoEligiblePackagesError) Error() string {
	switch {
	case e.AppName != "":
		return fmt.Sprintf("App '%s' has no eligible packages.", e.AppName)
	default:
		return fmt.Sprintf("No eligible packages available for app.")
	}

}
