package constant

// RelationshipType represetns the Cloud Controller To-One resource targetted
// by a relationship.
type RelationshipType string

const (
	// RelationshipTypeApplication is a relationship with a Cloud Controller
	// application.
	RelationshipTypeApplication RelationshipType = "app"

	// RelationshipTypeSpace is a relationship with a CloudController space.
	RelationshipTypeSpace RelationshipType = "space"

	// RelationshipTypeOrganization is a relationship with a CloudController organization.
	RelationshipTypeOrganization RelationshipType = "organization"
)
