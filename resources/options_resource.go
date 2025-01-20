package resources

import "strings"

func NewRouteOptions(options []string) map[string]*string {
	routeOptions := map[string]*string{}
	for _, option := range options {
		key, value, found := strings.Cut(option, "=")
		if found {
			routeOptions[key] = &value
		} else {
			routeOptions[option] = nil
		}
	}
	return routeOptions
}
