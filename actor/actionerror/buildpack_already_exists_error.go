package actionerror

import "fmt"

type BuildpackAlreadyExistsError string

func (e BuildpackAlreadyExistsError) Error() string {
	return fmt.Sprintf("A buildpack with the name %s already exists", string(e))
}
