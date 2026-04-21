package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) AddRoutePolicy(domainName, source, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}

	// Get the domain to ensure it exists and supports route policies
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

// GetRoutesByDomain gets routes for a domain with optional hostname and path filters
func (actor Actor) GetRoutesByDomain(domainGUID, hostname, path string) ([]resources.Route, Warnings, error) {
	queries := []ccv3.Query{
		{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
	}

	if hostname != "" {
		queries = append(queries, ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}})
	}

	if path != "" {
		queries = append(queries, ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}})
	}

	ccv3Routes, warnings, err := actor.CloudControllerClient.GetRoutes(queries...)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var routes []resources.Route
	for _, route := range ccv3Routes {
		routes = append(routes, resources.Route(route))
	}

	return routes, Warnings(warnings), nil
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

	// Build query for route policies filtered by space, with included routes
	queries := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.Include, Values: []string{"route"}},
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

	// Build route lookup map from included resources
	routeByGUID := make(map[string]resources.Route)
	for _, route := range includedResources.Routes {
		routeByGUID[route.GUID] = route
	}

	// Apply optional filters to the included routes
	if domainName != "" {
		domain, warnings, err := actor.GetDomainByName(domainName)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		// Filter routes by domain GUID
		filteredRoutes := make(map[string]resources.Route)
		for guid, route := range routeByGUID {
			if route.DomainGUID == domain.GUID {
				filteredRoutes[guid] = route
			}
		}
		routeByGUID = filteredRoutes
	}

	if hostname != "" {
		// Filter routes by hostname
		filteredRoutes := make(map[string]resources.Route)
		for guid, route := range routeByGUID {
			if route.Host == hostname {
				filteredRoutes[guid] = route
			}
		}
		routeByGUID = filteredRoutes
	}

	if path != "" {
		// Filter routes by path
		filteredRoutes := make(map[string]resources.Route)
		for guid, route := range routeByGUID {
			if route.Path == path {
				filteredRoutes[guid] = route
			}
		}
		routeByGUID = filteredRoutes
	}

	// Build domain name cache
	domainCache := make(map[string]string)
	for _, route := range routeByGUID {
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

	// Build results with route information and resolved sources
	// Only include route policies whose routes match the filters
	var results []RoutePolicyWithRoute
	for _, policy := range routePolicies {
		route, exists := routeByGUID[policy.RouteGUID]
		if !exists {
			// Skip policies for routes that don't match the optional filters
			continue
		}

		scopeType, sourceName, warnings, err := actor.resolveRoutePolicySource(policy.Source)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			// If we can't resolve the source, sourceName is already empty string
			// scopeType is still set correctly
		}

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

// resolveRoutePolicySource resolves a source to scope type and human-readable source name
func (actor Actor) resolveRoutePolicySource(source string) (scopeType string, sourceName string, warnings Warnings, err error) {
	allWarnings := Warnings{}

	// Parse source format: cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any
	if source == "cf:any" {
		return "any", "", nil, nil
	}

	// Split source into parts
	// Expected format: cf:type:guid
	const prefix = "cf:"
	if len(source) < len(prefix) {
		return "unknown", "", nil, nil
	}

	sourceBody := source[len(prefix):]
	parts := splitSource(sourceBody)
	if len(parts) < 2 {
		return "unknown", "", nil, nil
	}

	sourceType := parts[0]
	guid := parts[1]

	switch sourceType {
	case "app":
		apps, apiWarnings, err := actor.CloudControllerClient.GetApplications(
			ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{guid}},
		)
		allWarnings = append(allWarnings, Warnings(apiWarnings)...)
		if err != nil || len(apps) == 0 {
			return "app", "", allWarnings, err
		}
		return "app", apps[0].Name, allWarnings, nil

	case "space":
		spaces, _, apiWarnings, err := actor.CloudControllerClient.GetSpaces(
			ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{guid}},
		)
		allWarnings = append(allWarnings, Warnings(apiWarnings)...)
		if err != nil || len(spaces) == 0 {
			return "space", "", allWarnings, err
		}
		return "space", spaces[0].Name, allWarnings, nil

	case "org":
		orgs, apiWarnings, err := actor.CloudControllerClient.GetOrganizations(
			ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{guid}},
		)
		allWarnings = append(allWarnings, Warnings(apiWarnings)...)
		if err != nil || len(orgs) == 0 {
			return "org", "", allWarnings, err
		}
		return "org", orgs[0].Name, allWarnings, nil

	default:
		return "unknown", "", nil, nil
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
