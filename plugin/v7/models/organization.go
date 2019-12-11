// +build V7

package models

type Organization struct {
	// GUID is the unique organization identifier.
	GUID string
	// Name is the name of the organization.
	Name string

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata

	Spaces []Space
}
