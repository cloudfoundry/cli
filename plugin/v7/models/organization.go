// +build V7

package models

type OrgSummary struct {
	// GUID is the unique organization identifier.
	GUID string
	// Name is the name of the organization.
	Name string

	Spaces []Space
}
