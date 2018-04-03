package manifest

import "fmt"

type InvalidYAMLError struct {
	Err error
}

func (e InvalidYAMLError) Error() string {
	return fmt.Sprintf("The option --vars-file expects a valid YAML file. %s", e.Err)
}
