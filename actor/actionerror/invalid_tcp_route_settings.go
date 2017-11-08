package actionerror

import "fmt"

type InvalidTCPRouteSettings struct {
	Domain string
}

func (e InvalidTCPRouteSettings) Error() string {
	return fmt.Sprintln("Invalid TCP settings for", e.Domain)
}
