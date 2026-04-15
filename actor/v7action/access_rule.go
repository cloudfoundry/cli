package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) AddAccessRule(domainName, selector, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}

	// Get the domain to ensure it exists and supports access rules
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

	// Create the access rule
	accessRule := resources.AccessRule{
		Selector:  selector,
		RouteGUID: route.GUID,
	}

	_, apiWarnings, err := actor.CloudControllerClient.CreateAccessRule(accessRule)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetAccessRulesByRoute(domainName, hostname, path string) ([]resources.AccessRule, Warnings, error) {
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

	// Get access rules for this route
	accessRules, _, apiWarnings, err := actor.CloudControllerClient.GetAccessRules(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{route.GUID}},
	)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)

	var rules []resources.AccessRule
	for _, rule := range accessRules {
		rules = append(rules, resources.AccessRule(rule))
	}

	return rules, allWarnings, err
}

func (actor Actor) DeleteAccessRuleBySelector(domainName, selector, hostname, path string) (Warnings, error) {
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

	// Get access rules for this route to find the one with matching selector
	accessRules, _, apiWarnings, err := actor.CloudControllerClient.GetAccessRules(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{route.GUID}},
	)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	// Find the rule with matching selector
	var ruleGUID string
	for _, rule := range accessRules {
		if rule.Selector == selector {
			ruleGUID = rule.GUID
			break
		}
	}

	if ruleGUID == "" {
		return allWarnings, actionerror.AccessRuleNotFoundError{Selector: selector}
	}

	// Delete the access rule
	_, deleteWarnings, err := actor.CloudControllerClient.DeleteAccessRule(ruleGUID)
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

// AccessRuleWithRoute combines an access rule with its associated route information
type AccessRuleWithRoute struct {
	resources.AccessRule
	Route      resources.Route
	DomainName string
	ScopeType  string // "app", "space", "org", or "any"
	SourceName string // Resolved source name (app/space/org) or empty string
}

// GetAccessRulesForSpace gets all access rules for routes in a space with optional filters
func (actor Actor) GetAccessRulesForSpace(
	spaceGUID string,
	domainName string,
	hostname string,
	path string,
	labelSelector string,
) ([]AccessRuleWithRoute, Warnings, error) {
	allWarnings := Warnings{}

	// Build query for access rules filtered by space, with included routes
	queries := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.Include, Values: []string{"route"}},
	}

	// Add label selector if provided
	if labelSelector != "" {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	// Fetch access rules directly by space GUID with included routes (single API call)
	accessRules, includedResources, apiWarnings, err := actor.CloudControllerClient.GetAccessRules(queries...)
	allWarnings = append(allWarnings, Warnings(apiWarnings)...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(accessRules) == 0 {
		// No access rules found - return empty list, not an error
		return []AccessRuleWithRoute{}, allWarnings, nil
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
	// Only include access rules whose routes match the filters
	var results []AccessRuleWithRoute
	for _, rule := range accessRules {
		route, exists := routeByGUID[rule.RouteGUID]
		if !exists {
			// Skip rules for routes that don't match the optional filters
			continue
		}

		scopeType, sourceName, warnings, err := actor.resolveAccessRuleSource(rule.Selector)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			// If we can't resolve the source, sourceName is already empty string
			// scopeType is still set correctly
		}

		results = append(results, AccessRuleWithRoute{
			AccessRule: resources.AccessRule(rule),
			Route:      route,
			DomainName: domainCache[route.DomainGUID],
			ScopeType:  scopeType,
			SourceName: sourceName,
		})
	}

	return results, allWarnings, nil
}

// resolveAccessRuleSource resolves a selector to scope type and human-readable source name
func (actor Actor) resolveAccessRuleSource(selector string) (scopeType string, sourceName string, warnings Warnings, err error) {
	allWarnings := Warnings{}

	// Parse selector format: cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any
	if selector == "cf:any" {
		return "any", "", nil, nil
	}

	// Split selector into parts
	// Expected format: cf:type:guid
	const prefix = "cf:"
	if len(selector) < len(prefix) {
		return "unknown", "", nil, nil
	}

	selectorBody := selector[len(prefix):]
	parts := splitSelector(selectorBody)
	if len(parts) < 2 {
		return "unknown", "", nil, nil
	}

	selectorType := parts[0]
	guid := parts[1]

	switch selectorType {
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

// splitSelector splits a selector body by colon, handling the case where
// the selector might be "type:guid" format
func splitSelector(s string) []string {
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
