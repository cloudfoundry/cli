package resources

import "strings"

func CreateRouteOptions(options []string) (map[string]*string, []string) {
	routeOptions := map[string]*string{}
	wrongOptSpecs := []string{}
	for _, option := range options {
		key, value, found := strings.Cut(option, "=")
		if found {
			routeOptions[key] = &value
		} else {
			wrongOptSpecs = append(wrongOptSpecs, option)
		}
	}
	if len(wrongOptSpecs) == 0 {
		return routeOptions, nil
	}
	return routeOptions, wrongOptSpecs
}

func RemoveRouteOptions(options []string) map[string]*string {
	routeOptions := map[string]*string{}
	for _, option := range options {
		routeOptions[option] = nil
	}
	return routeOptions
}
