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
	// AvailableFilter is a query parameter for listing available resources
	AvailableFilter QueryKey = "available"
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
	// RouteGUIDFilter is a query parameter for listing objects by Route GUID.
	RouteGUIDFilter QueryKey = "route_guids"
	// ServiceInstanceGUIDFilter is a query parameter for listing objects by Service Instance GUID.
	ServiceInstanceGUIDFilter QueryKey = "service_instance_guids"
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
	// PortFilter is a query param for getting an object with the given port (TCP routes)
	PortFilter QueryKey = "port"
	// PortsFilter is a query param for getting an object with the given ports (TCP routes)
	PortsFilter QueryKey = "ports"
	// RoleTypesFilter is a query param for getting a role by type
	RoleTypesFilter QueryKey = "types"
	// StackFilter is a query parameter for listing objects by stack name
	StackFilter QueryKey = "stacks"
	// TypeFiler is a query parameter for selecting binding type
	TypeFilter QueryKey = "type"
	// UnmappedFilter is a query parameter specifying unmapped routes
	UnmappedFilter QueryKey = "unmapped"
	// UserGUIDFilter is a query parameter when getting a user by GUID
	UserGUIDFilter QueryKey = "user_guids"
	// UsernamesFilter is a query parameter when getting a user by username
	UsernamesFilter QueryKey = "usernames"
	// StatesFilter is a query parameter when getting a package's droplets by state
	VersionsFilter QueryKey = "versions"
	// VersionsFilter is a query parameter when getting an apps revisions by version
	StatesFilter QueryKey = "states"
	// ServiceBrokerNamesFilter is a query parameter when getting plans or offerings according to the Service Brokers that it relates to
	ServiceBrokerNamesFilter QueryKey = "service_broker_names"
	// ServiceBrokerGUIDsFilter is a query parameter for getting resources according to the service broker GUID
	ServiceBrokerGUIDsFilter QueryKey = "service_broker_guids"
	// ServiceOfferingNamesFilter is a query parameter when getting a plan according to the Service Offerings that it relates to
	ServiceOfferingNamesFilter QueryKey = "service_offering_names"
	// ServiceOfferingGUIDsFilter is a query parameter when getting resources according to service offering GUIDs
	ServiceOfferingGUIDsFilter QueryKey = "service_offering_guids"
	// FieldsServiceOfferingServiceBroker is a query parameter to include specific fields from a service broker in a plan response
	FieldsServiceOfferingServiceBroker QueryKey = "fields[service_offering.service_broker]"
	// FieldsServiceBroker is a query parameter to include specific fields from a service broker in an offering response
	FieldsServiceBroker QueryKey = "fields[service_broker]"
	// FieldsServicePlan is a query parameter to include specific fields from a service plan
	FieldsServicePlan QueryKey = "fields[service_plan]"
	// FieldsServicePlanServiceOffering is a query parameter to include specific fields from a service offering
	FieldsServicePlanServiceOffering QueryKey = "fields[service_plan.service_offering]"
	// FieldsServicePlanServiceOfferingServiceBroker is a query parameter to include specific fields from a service broker
	FieldsServicePlanServiceOfferingServiceBroker QueryKey = "fields[service_plan.service_offering.service_broker]"
	// FieldsSpace is a query parameter to include specific fields from a space
	FieldsSpace QueryKey = "fields[space]"
	// FieldsSpaceOrganization is a query parameter to include specific fields from a organization
	FieldsSpaceOrganization QueryKey = "fields[space.organization]"

	// OrderBy is a query parameter to specify how to order objects.
	OrderBy QueryKey = "order_by"
	// PerPage is a query parameter for specifying the number of results per page.
	PerPage QueryKey = "per_page"
	// Page is a query parameter for specifying the number of the requested page.
	Page QueryKey = "page"
	// Include is a query parameter for specifying other resources associated with the
	// resource returned by the endpoint
	Include QueryKey = "include"

	// GloballyEnabledStaging is the query parameter for getting only security groups that are globally enabled for staging
	GloballyEnabledStaging QueryKey = "globally_enabled_staging"

	// GloballyEnabledRunning is the query parameter for getting only security groups that are globally enabled for running
	GloballyEnabledRunning QueryKey = "globally_enabled_running"

	// NameOrder is a query value for ordering by name. This value is used in
	// conjunction with the OrderBy QueryKey.
	NameOrder = "name"

	// PositionOrder is a query value for ordering by position. This value is
	// used in conjunction with the OrderBy QueryKey.
	PositionOrder = "position"

	// CreatedAtDescendingOrder is a query value for ordering by created_at timestamp,
	// in descending order.
	CreatedAtDescendingOrder = "-created_at"

	// SourceGUID is the query parameter for getting an object. Currently it's used as a package GUID
	// to retrieve a package to later copy it to an app (CopyPackage())
	SourceGUID = "source_guid"

	// Purge is a query parameter used on a Delete request to indicate that dependent resources should also be deleted
	Purge = "purge"

	// MaxPerPage is the largest value of "per_page" that should be used
	MaxPerPage = "5000"
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
		if query.Key == NameFilter {
			encodedParamValues := []string{}
			for _, valString := range query.Values {
				commaEncoded := strings.ReplaceAll(valString, ",", "%2C")
				encodedParamValues = append(encodedParamValues, commaEncoded)
			}
			query.Values = encodedParamValues
		}

		params.Add(string(query.Key), strings.Join(query.Values, ","))
	}

	return params
}
