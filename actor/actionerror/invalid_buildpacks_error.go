package actionerror

import "fmt"

type InvalidBuildpacksError struct {
}

func (err InvalidBuildpacksError) Error() string {
	return fmt.Sprintf("Multiple buildpacks flags cannot have null/default option.")
}
