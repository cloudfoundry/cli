package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) AddRoutePolicy(domainName, source, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}

	// Get the domain and verify it has route policy enforcement enabled
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if !domain.EnforceRoutePolicies.IsSet || !domain.EnforceRoutePolicies.Value {
		return allWarnings, actionerror.DomainNotEnforcingRoutePoliciesError{Name: domainName}
	}

	// Find the route
	routes, routeWarnings, err := actor.GetRoutesByDomain(domain.GUID, hostname, path)
	allWarnings = append(allWarnings, routeWarnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(routes) == 0 {
		return allWarnings, actionerror.RouteNotFoundError{
			Host:       hostname,
			DomainName: domainName,
			Path:       path,
		}
	}

	route := routes[0]

	// Create the route policy
	routePolicy := resources.RoutePolicy{
		Source:    source,
		RouteGUID: route.GUID,
	}

	_, apiWarnings, err := actor.CloudControllerClient.CreateRoutePolicy(routePolicy)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetRoutePoliciesByRoute(domainName, hostname, path string) ([]resources.RoutePolicy, Warnings, error) {
	allWarnings := Warnings{}

	// Get the domain
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	// Find the route
	routes, routeWarnings, err := actor.GetRoutesByDomain(domain.GUID, hostname, path)
	allWarnings = append(allWarnings, routeWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(routes) == 0 {
		return nil, allWarnings, actionerror.RouteNotFoundError{
			Host:       hostname,
			DomainName: domainName,
			Path:       path,
		}
	}

	route := routes[0]

	// Get route policies for this route
	routePolicies, _, apiWarnings, err := actor.CloudControllerClient.GetRoutePolicies(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{route.GUID}},
	)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)

	var policies []resources.RoutePolicy
	for _, policy := range routePolicies {
		policies = append(policies, resources.RoutePolicy(policy))
	}

	return policies, allWarnings, err
}

func (actor Actor) DeleteRoutePolicyBySource(domainName, source, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}

	// Get the domain
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	// Find the route
	routes, routeWarnings, err := actor.GetRoutesByDomain(domain.GUID, hostname, path)
	allWarnings = append(allWarnings, routeWarnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(routes) == 0 {
		return allWarnings, actionerror.RouteNotFoundError{
			Host:       hostname,
			DomainName: domainName,
			Path:       path,
		}
	}

	route := routes[0]

	// Get route policies for this route to find the one with matching source
	routePolicies, _, apiWarnings, err := actor.CloudControllerClient.GetRoutePolicies(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{route.GUID}},
	)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	// Find the policy with matching source
	var policyGUID string
	for _, policy := range routePolicies {
		if policy.Source == source {
			policyGUID = policy.GUID
			break
		}
	}

	if policyGUID == "" {
		return allWarnings, actionerror.RoutePolicyNotFoundError{Source: source}
	}

	// Delete the route policy
	_, deleteWarnings, err := actor.CloudControllerClient.DeleteRoutePolicy(policyGUID)
	allWarnings = append(allWarnings, Warnings(deleteWarnings)...)

	return allWarnings, err
}


// RoutePolicyWithRoute combines a route policy with its associated route information
type RoutePolicyWithRoute struct {
	resources.RoutePolicy
	Route      resources.Route
	DomainName string
	ScopeType  string // "app", "space", "org", or "any"
	SourceName string // Resolved source name (app/space/org) or empty string
}

// GetRoutePoliciesForSpace gets all route policies for routes in a space with optional filters
func (actor Actor) GetRoutePoliciesForSpace(
	spaceGUID string,
	domainName string,
	hostname string,
	path string,
	labelSelector string,
) ([]RoutePolicyWithRoute, Warnings, error) {
	allWarnings := Warnings{}

	// Build query for route policies filtered by space, with included routes and sources.
	// ?include=route,source causes CAPI to return all referenced apps/spaces/orgs inline,
	// avoiding per-policy follow-up lookups.
	queries := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.Include, Values: []string{"route,source"}},
	}

	// Add label selector if provided
	if labelSelector != "" {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	// Fetch route policies directly by space GUID with included routes (single API call)
	routePolicies, includedResources, apiWarnings, err := actor.CloudControllerClient.GetRoutePolicies(queries...)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(routePolicies) == 0 {
		// No route policies found - return empty list, not an error
		return []RoutePolicyWithRoute{}, allWarnings, nil
	}

	// Create domain name cache; pre-populated below if a domain filter is used.
	domainCache := make(map[string]string)

	// Filter the routes slice before building the lookup map.
	filteredRoutes := includedResources.Routes

	if domainName != "" {
		domain, warnings, err := actor.GetDomainByName(domainName)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		// Pre-populate cache: all matching routes share this single domain GUID.
		domainCache[domain.GUID] = domain.Name
		var matched []resources.Route
		for _, route := range filteredRoutes {
			if route.DomainGUID == domain.GUID {
				matched = append(matched, route)
			}
		}
		filteredRoutes = matched
	}

	if hostname != "" {
		var matched []resources.Route
		for _, route := range filteredRoutes {
			if route.Host == hostname {
				matched = append(matched, route)
			}
		}
		filteredRoutes = matched
	}

	if path != "" {
		var matched []resources.Route
		for _, route := range filteredRoutes {
			if route.Path == path {
				matched = append(matched, route)
			}
		}
		filteredRoutes = matched
	}

	// Build route lookup map from filtered routes only.
	routeByGUID := make(map[string]resources.Route, len(filteredRoutes))
	for _, route := range filteredRoutes {
		routeByGUID[route.GUID] = route
	}

	// Populate domain cache for any domain GUIDs not yet known.
	for _, route := range filteredRoutes {
		if _, exists := domainCache[route.DomainGUID]; !exists {
			domain, warnings, err := actor.GetDomain(route.DomainGUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				// If we can't get the domain, use the GUID as fallback
				domainCache[route.DomainGUID] = route.DomainGUID
			} else {
				domainCache[route.DomainGUID] = domain.Name
			}
		}
	}

	// Build name lookup maps from included source resources.
	// CAPI returns all referenced apps/spaces/orgs inline when ?include=source is sent,
	// so no additional API calls are required per policy.
	appNameByGUID := make(map[string]string)
	for _, app := range includedResources.Apps {
		appNameByGUID[app.GUID] = app.Name
	}
	spaceNameByGUID := make(map[string]string)
	for _, space := range includedResources.Spaces {
		spaceNameByGUID[space.GUID] = space.Name
	}
	orgNameByGUID := make(map[string]string)
	for _, org := range includedResources.Organizations {
		orgNameByGUID[org.GUID] = org.Name
	}

	// Build results with route information and resolved sources.
	// Only include route policies whose routes match the filters.
	var results []RoutePolicyWithRoute
	for _, policy := range routePolicies {
		route, exists := routeByGUID[policy.RouteGUID]
		if !exists {
			// Skip policies for routes that don't match the optional filters
			continue
		}

		scopeType, sourceName := sourceInfoFromIncluded(policy.Source, appNameByGUID, spaceNameByGUID, orgNameByGUID)

		results = append(results, RoutePolicyWithRoute{
			RoutePolicy: resources.RoutePolicy(policy),
			Route:       route,
			DomainName:  domainCache[route.DomainGUID],
			ScopeType:   scopeType,
			SourceName:  sourceName,
		})
	}

	return results, allWarnings, nil
}

// sourceInfoFromIncluded resolves a source string to a scope type and human-readable name
// using the name maps pre-built from CAPI's ?include=source response.
// It performs no API calls.
func sourceInfoFromIncluded(source string, apps, spaces, orgs map[string]string) (scopeType, sourceName string) {
	if source == "cf:any" {
		return "any", ""
	}

	const prefix = "cf:"
	if len(source) < len(prefix) {
		return "unknown", ""
	}

	parts := splitSource(source[len(prefix):])
	if len(parts) < 2 {
		return "unknown", ""
	}

	sourceType, guid := parts[0], parts[1]
	switch sourceType {
	case "app":
		return "app", apps[guid]
	case "space":
		return "space", spaces[guid]
	case "org":
		return "org", orgs[guid]
	default:
		return "unknown", ""
	}
}

// splitSource splits a source body by colon, handling the case where
// the source might be "type:guid" format
func splitSource(s string) []string {
	var parts []string
	current := ""
	for _, char := range s {
		if char == ':' && len(parts) == 0 {
			// First colon - split here
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
