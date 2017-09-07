package ccv3

const (
	// GUIDFilter is a query paramater for listing objects by GUID.
	GUIDFilter = "guids"
	// NameFilter is a query paramater for listing objects by name.
	NameFilter = "names"
	// AppGUIDFilter is a query paramater for listing objects by app GUID.
	AppGUIDFilter = "app_guids"
	// OrganizationGUIDFilter is a query paramater for listing objects by Organization GUID.
	OrganizationGUIDFilter = "organization_guids"
	// SpaceGUIDFilter is a query paramater for listing objects by Space GUID.
	SpaceGUIDFilter = "space_guids"

	// OrderBy is a query paramater to specify how to order objects.
	OrderBy = "order_by"
	// NameOrder is value for a query paramater when ordering by name.
	NameOrder = "name"
)
