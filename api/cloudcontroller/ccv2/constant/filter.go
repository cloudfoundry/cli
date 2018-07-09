package constant

// FilterType is the type of filter a Filter uses.
type FilterType string

const (
	// AppGUIDFilter is the name of the 'app_guid' filter.
	AppGUIDFilter FilterType = "app_guid"
	// DomainGUIDFilter is the name of the 'domain_guid' filter.
	DomainGUIDFilter FilterType = "domain_guid"
	// OrganizationGUIDFilter is the name of the 'organization_guid' filter.
	OrganizationGUIDFilter FilterType = "organization_guid"
	// RouteGUIDFilter is the name of the 'route_guid' filter.
	RouteGUIDFilter FilterType = "route_guid"
	// ServiceBrokerGUIDFilter is the name of the 'service_broker_guid' filter.
	ServiceBrokerGUIDFilter FilterType = "service_broker_guid"
	// ServiceGUIDFilter is the name of the 'service_guid' filter.
	ServiceGUIDFilter FilterType = "service_guid"
	// ServiceInstanceGUIDFilter is the name of the 'service_instance_guid' filter.
	ServiceInstanceGUIDFilter FilterType = "service_instance_guid"
	// ServicePlanGUIDFilter is the name of the 'service_plan_guid' filter.
	ServicePlanGUIDFilter FilterType = "service_plan_guid"
	// SpaceGUIDFilter is the name of the 'space_guid' filter.
	SpaceGUIDFilter FilterType = "space_guid"

	// LabelFilter is the name of the 'label' filter.
	LabelFilter FilterType = "label"
	// NameFilter is the name of the 'name' filter.
	NameFilter FilterType = "name"
	// HostFilter is the name of the 'host' filter.
	HostFilter FilterType = "host"
	// PathFilter is the name of the 'path' filter.
	PathFilter FilterType = "path"
	// PortFilter is the name of the 'port' filter.
	PortFilter FilterType = "port"

	// TimestampFilter is the name of the 'timestamp' filter.
	TimestampFilter FilterType = "timestamp"
	// TypeFilter is the name of the 'type' filter.
	TypeFilter FilterType = "type"
)

// FilterOperator is the type of operation a Filter uses.
type FilterOperator string

const (
	// EqualOperator is the Filter's equal operator.
	EqualOperator FilterOperator = ":"

	// GreaterThanOperator is the query greater than operator.
	GreaterThanOperator FilterOperator = ">"

	// InOperator is the Filter's "IN" operator.
	InOperator FilterOperator = " IN "
)
