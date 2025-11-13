package resources

import "code.cloudfoundry.org/cli/v8/types"

type Sidecar struct {
	GUID    string               `json:"guid"`
	Name    string               `json:"name"`
	Command types.FilteredString `json:"command"`
}
