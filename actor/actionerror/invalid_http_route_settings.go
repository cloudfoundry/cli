package actionerror

import "fmt"

type InvalidHTTPRouteSettings struct {
	Domain string
}

func (e InvalidHTTPRouteSettings) Error() string {
	return fmt.Sprintln("Invalid HTTP settings for", e.Domain)
}
