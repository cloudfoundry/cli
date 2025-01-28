package resources

import "strings"

func CreateRouteOptions(options []string) (map[string]*string, *string) {
	routeOptions := map[string]*string{}
	for _, option := range options {
		key, value, found := strings.Cut(option, "=")
		if found {
			routeOptions[key] = &value
		} else {
			return routeOptions, &option
		}
	}
	return routeOptions, nil
}

func RemoveRouteOptions(options []string) map[string]*string {
	routeOptions := map[string]*string{}
	for _, option := range options {
		routeOptions[option] = nil
	}
	return routeOptions
}
