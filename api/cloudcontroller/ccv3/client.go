// Package ccv3 represents a Cloud Controller V3 client.
//
// It is currently designed to support Cloud Controller API 2.64.0. However, it
// may include features and endpoints of later API versions.
//
// For more information on the Cloud Controller API see
// https://apidocs.cloudfoundry.org/
//
// Method Naming Conventions
//
// The client takes a '<Action Name><Top Level Endpoint><Return Value>'
// approach to method names.  If the <Top Level Endpoint> and <Return Value>
// are similar, they do not need to be repeated. If a GUID is required for the
// <Top Level Endpoint>, the pluralization is removed from said endpoint in the
// method name.
//
// For Example:
//   Method Name: GetApplication
//   Endpoint: /v3/applications/:guid
//   Action Name: Get
//   Top Level Endpoint: applications
//   Return Value: Application
//
//   Method Name: GetServiceInstances
//   Endpoint: /v3/service_instances
//   Action Name: Get
//   Top Level Endpoint: service_instances
//   Return Value: []ServiceInstance
//
//   Method Name: GetSpaceServiceInstances
//   Endpoint: /v3/spaces/:guid/service_instances
//   Action Name: Get
//   Top Level Endpoint: spaces
//   Return Value: []ServiceInstance
//
// Use the following table to determine which HTTP Command equates to which
// Action Name:
//   HTTP Command -> Action Name
//   POST -> New
//   GET -> Get
//   PUT -> Update
//   DELETE -> Delete
//
// Method Locations
//
// Methods exist in the same file as their return type, regardless of which
// endpoint they use.
package ccv3

import (
	"fmt"
	"runtime"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// Warnings are a collection of warnings that the Cloud Controller can return
// back from an API request.
type Warnings []string

// Client can be used to talk to a Cloud Controller's V3 Endpoints.
type Client struct {
	APIInfo
	cloudControllerURL string

	connection cloudcontroller.Connection
	userAgent  string
}

// NewClient returns a new Client.
func NewClient(appName string, appVersion string) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)", appName, appVersion, runtime.Version(), runtime.GOARCH, runtime.GOOS)
	return &Client{
		userAgent: userAgent,
	}
}
