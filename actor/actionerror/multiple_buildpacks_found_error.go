package actionerror

import (
	"fmt"
)

// MultipleBuildpacksFoundError represents the scenario when the cloud
// controller returns multiple buildpacks when filtering by name.
type MultipleBuildpacksFoundError struct {
	BuildpackName string
}

func (e MultipleBuildpacksFoundError) Error() string {
	return fmt.Sprintf("Multiple buildpacks named %s found", e.BuildpackName)
}
