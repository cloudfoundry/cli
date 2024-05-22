// Package ccv2 represents a Cloud Controller V2 client.
//
// These sets of packages are still under development/pre-pre-pre...alpha. Use
// at your own risk! Functionality and design may change without warning.
//
// It is currently designed to support Cloud Controller API 2.58.0. However, it
// may include features and endpoints of later API versions.
//
// For more information on the Cloud Controller API see
// https://apidocs.cloudfoundry.org/
//
// # Method Naming Conventions
//
// The client takes a '<Action Name><Top Level Endpoint><nth... level
// endpoint><Return Value>' approach to method names.  If the <Top Level
// Endpoint> and <Return Value> are similar, they do not need to be repeated.
// If a GUID is required for the <Top Level Endpoint>, the pluralization is
// removed from said endpoint in the method name.
//
// For Example:
//
//	Method Name: GetApplication
//	Endpoint: /v2/applications/:guid
//	Action Name: GET
//	Top Level Endpoint: applications
//	Return Value: Application
//
//	Method Name: GetServiceInstances
//	Endpoint: /v2/service_instances
//	Action Name: GET
//	Top Level Endpoint: service_instances
//	Return Value: []ServiceInstance
//
//	Method Name: GetSpaceServiceInstances
//	Endpoint: /v2/spaces/:guid/service_instances
//	Action Name: GET
//	Top Level Endpoint: spaces
//	Return Value: []ServiceInstance
//
//	Method Name: UpdateRouteApplication
//	Endpoint: /v2/routes/:route_guid/apps/:app_guid
//	Action Name: PUT
//	Top Level Endpoint: routes
//	Second Level Endpoint: Application
//	Return Value: Route
//
//	Method Name: DeleteSpaceJob
//	Endpoint: /v2/spaces/:space_guid
//	Action Name: DELETE
//	Top Level Endpoint: spaces
//	Return Value: Job
//
// Use the following table to determine which HTTP Command equates to which
// Action Name:
//
//	HTTP Command -> Action Name
//	POST -> Create
//	GET -> Get
//	PUT -> Update
//	DELETE -> Delete
//	PATCH -> Patch
//
// # Method Locations
//
// Methods exist in the same file as their return type, regardless of which
// endpoint they use.
//
// # Error Handling
//
// All error handling that requires parsing the error_code/code returned back
// from the Cloud Controller should be placed in the errorWrapper. Everything
// else can be handled in the individual operations. All errors structs should
// be placed in individual files in the ccerror package.
//
// # No inline-relations-depth And summary Endpoints
//
// This package will not ever use 'inline-relations-depth' or the
// '/summary' endpoints for any operations. These requests can be extremely
// taxing on the Cloud Controller and are avoided at all costs. Additionally,
// the objects returned back from these requests can become extremely
// inconsistent across versions and are problematic to deal with in general.
//
// An additional problem that occurs with these endpoints is that lists of 51+
// objects do not get returned by the CC. Example: Summary of Organizations
// includes a list of spaces. If there are 50 spaces, the endpoint returns
// these spaces under the 'spaces' field; if there are 51+, spaces, the
// 'spaces' field is missing from the CC return.
package ccv2

import (
	"fmt"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"github.com/tedsuo/rata"
)

// Warnings are a collection of warnings that the Cloud Controller can return
// back from an API request.
type Warnings []string

// Client is a client that can be used to talk to a Cloud Controller's V2
// Endpoints.
type Client struct {
	authorizationEndpoint     string
	cloudControllerAPIVersion string
	cloudControllerURL        string
	dopplerEndpoint           string
	logCacheEndpoint          string
	minCLIVersion             string
	routingEndpoint           string

	jobPollingInterval time.Duration
	jobPollingTimeout  time.Duration

	connection cloudcontroller.Connection
	router     *rata.RequestGenerator
	userAgent  string
	wrappers   []ConnectionWrapper
}

// Config allows the Client to be configured
type Config struct {
	// AppName is the name of the application/process using the client.
	AppName string

	// AppVersion is the version of the application/process using the client.
	AppVersion string

	// JobPollingTimeout is the maximum amount of time a job polls for.
	JobPollingTimeout time.Duration

	// JobPollingInterval is the wait time between job polls.
	JobPollingInterval time.Duration

	// Wrappers that apply to the client connection.
	Wrappers []ConnectionWrapper
}

// NewClient returns a new Cloud Controller Client.
func NewClient(config Config) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)", config.AppName, config.AppVersion, runtime.Version(), runtime.GOARCH, runtime.GOOS)
	return &Client{
		userAgent:          userAgent,
		jobPollingInterval: config.JobPollingInterval,
		jobPollingTimeout:  config.JobPollingTimeout,
		wrappers:           append([]ConnectionWrapper{newErrorWrapper()}, config.Wrappers...),
	}
}
