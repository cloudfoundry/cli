package actionerror

import (
	"fmt"
	"sort"
	"strings"
)

type AmbiguousUserError struct {
	Username string
	Origins  []string
}

func (e AmbiguousUserError) Error() string {
	sort.Strings(e.Origins)
	origins := strings.Join(e.Origins, ", ")
	return fmt.Sprintf(
		"Ambiguous user. User with username '%s' exists in the following origins: %s. Specify an origin to disambiguate.",
		e.Username,
		origins,
	)
}
