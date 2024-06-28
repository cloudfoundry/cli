package constant

// RelationshipType represents the Cloud Controller To-One resource targeted by
// a relationship.
type RelationshipType string

const (
	// RelationshipTypeApplication is a relationship with a Cloud Controller
	// application.
	RelationshipTypeApplication RelationshipType = "app"

	// RelationshipTypeSpace is a relationship with a Cloud Controller space.
	RelationshipTypeSpace RelationshipType = "space"

	// RelationshipTypeOrganization is a relationship with a Cloud Controller
	// organization.
	RelationshipTypeOrganization RelationshipType = "organization"

	// RelationshipTypeUser is a relationship with a Cloud Controller user.
	RelationshipTypeUser RelationshipType = "user"

	// RelationshipTypeQuota is a relationship with a Cloud Controller quota (org quota or space quota).
	RelationshipTypeQuota RelationshipType = "quota"

	// RelationshipTypeCurrentDroplet is a relationship with a Droplet.
	RelationshipTypeCurrentDroplet RelationshipType = "current_droplet"
)
