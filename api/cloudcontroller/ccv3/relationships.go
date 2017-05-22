package ccv3

type RelationshipType string

const (
	ApplicationRelationship RelationshipType = "app"
	SpaceRelationship       RelationshipType = "space"
)

// Relationships is a map of RelationshipTypes to Relationship.
type Relationships map[RelationshipType]Relationship
