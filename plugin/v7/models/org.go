// +build V7

package models

// Org represents a Cloud Controller V3 Org.
type Org struct {
	Name string
	GUID string
}

type Domain struct {
	Name string
	GUID string
}

type OrgSummary struct {
	Org

	Metadata *Metadata

	Spaces []Space

	Domains []Domain
}
