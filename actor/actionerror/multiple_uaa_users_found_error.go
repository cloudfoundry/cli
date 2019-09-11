package actionerror

import (
	"fmt"
	"strings"
)

type MultipleUAAUsersFoundError struct {
	Username string
	Origins  []string
}

func (e MultipleUAAUsersFoundError) Error() string {
	origins := strings.Join(e.Origins, ", ")
	return fmt.Sprintf("The username '%s' is found in multiple origins: %s.", e.Username, origins)
}
