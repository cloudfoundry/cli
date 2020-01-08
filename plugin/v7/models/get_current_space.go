// +build V7

package models

// Space represents a Cloud Controller V3 Space.
type CurrentSpace struct {
	Name string
	GUID string
}

type Space struct {
	Name     string
	GUID     string
	Metadata Metadata
}
