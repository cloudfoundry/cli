package internal

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/internal"
)

// Naming convention:
//
// Method + non-parameter parts of the path
//
// If the request returns a single entity by GUID, use the singular (for example
// /v2/organizations/:organization_guid is GetOrganization).
//
// The const name should always be the const value + Request.
const (
	GetRouterGroups = "GetRouterGroups"
)

// APIRoutes is a list of routes used by the rata library to construct request
// URLs.
var APIRoutes = map[string]internal.Route{
	GetRouterGroups: {Path: "/v1/router_groups", Method: http.MethodGet},
}
