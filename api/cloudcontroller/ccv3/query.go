package ccv3

import (
	"net/url"
	"strings"
)

// QueryKey is the type of query that is being selected on.
type QueryKey string

const (
	// AppGUIDFilter is a query parameter for listing objects by app GUID.
	AppGUIDFilter QueryKey = "app_guids"
	// GUIDFilter is a query parameter for listing objects by GUID.
	GUIDFilter QueryKey = "guids"
	// LabelSelectorFilter is a query parameter for listing objects by label
	LabelSelectorFilter QueryKey = "label_selector"
	// NameFilter is a query parameter for listing objects by name.
	NameFilter QueryKey = "names"
	// NoRouteFilter is a query parameter for skipping route creation and unmapping existing routes.
	NoRouteFilter QueryKey = "no_route"
	// OrganizationGUIDFilter is a query parameter for listing objects by Organization GUID.
	OrganizationGUIDFilter QueryKey = "organization_guids"
	// SequenceIDFilter is a query parameter for listing objects by sequence ID.
	SequenceIDFilter QueryKey = "sequence_ids"
	// SpaceGUIDFilter is a query parameter for listing objects by Space GUID.
	SpaceGUIDFilter QueryKey = "space_guids"
	// StatusValueFilter is a query parameter for listing deployments by status.value
	StatusValueFilter QueryKey = "status_values"
	// DomainGUIDFilter is a query param for listing events by target_guid
	TargetGUIDFilter QueryKey = "target_guids"
	// DomainGUIDFilter is a query param for listing objects by domain_guid
	DomainGUIDFilter QueryKey = "domain_guids"
	// HostsFilter is a query param for listing objects by hostname
	HostsFilter QueryKey = "hosts"
	// HostFilter is a query param for getting an object with the given host
	HostFilter QueryKey = "host"
	// Origins filter is a query parameter when getting a user by origin (Note: CAPI will return an error if usernames filter is not also provided)
	OriginsFilter QueryKey = "origins"
	// PathsFilter is a query param for listing objects by path
	PathsFilter QueryKey = "paths"
	// PathFilter is a query param for getting an object with the given host
	PathFilter QueryKey = "path"
	// RoleTypesFilter is a query param for getting a role by type
	RoleTypesFilter QueryKey = "types"
	// StackFilter is a query parameter for listing objects by stack name
	StackFilter QueryKey = "stacks"
	// UnmappedFilter is a query parameter specifying unmapped routes
	UnmappedFilter QueryKey = "unmapped"
	// UserGUIDFilter is a query parameter when getting a user by GUID
	UserGUIDFilter QueryKey = "user_guids"
	// UsernamesFilter is a query parameter when getting a user by username
	UsernamesFilter QueryKey = "usernames"
	// StatesFilter is a query parameter when getting a package's droplets by state
	StatesFilter QueryKey = "states"
	// ServiceBrokerNamesFilter is a query parameter when getting a resource according to the Service Brokers that it relates to
	ServiceBrokerNamesFilter QueryKey = "service_broker_names"

	// OrderBy is a query parameter to specify how to order objects.
	OrderBy QueryKey = "order_by"
	// PerPage is a query parameter for specifying the number of results per page.
	PerPage QueryKey = "per_page"
	// Include is a query parameter for specifying other resources associated with the
	// resource returned by the endpoint
	Include QueryKey = "include"

	// NameOrder is a query value for ordering by name. This value is used in
	// conjunction with the OrderBy QueryKey.
	NameOrder = "name"

	// PositionOrder is a query value for ordering by position. This value is
	// used in conjunction with the OrderBy QueryKey.
	PositionOrder = "position"

	// CreatedAtDescendingOrder is a query value for ordering by created_at timestamp,
	// in descending order.
	CreatedAtDescendingOrder = "-created_at"
)

// Query is additional settings that can be passed to some requests that can
// filter, sort, etc. the results.
type Query struct {
	Key    QueryKey
	Values []string
}

// FormatQueryParameters converts a Query object into a collection that
// cloudcontroller.Request can accept.
func FormatQueryParameters(queries []Query) url.Values {
	params := url.Values{}
	for _, query := range queries {
		params.Add(string(query.Key), strings.Join(query.Values, ","))
	}

	return params
}
