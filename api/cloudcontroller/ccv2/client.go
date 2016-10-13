// Package ccv2 represents a Cloud Controller V2 client.
//
// It is currently designed to support Cloud Controller API 2.29.0. However, it
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
//   Endpoint: /v2/applications/:guid
//   Action Name: Get
//   Top Level Endpoint: applications
//   Return Value: Application
//
//   Method Name: GetServiceInstances
//   Endpoint: /v2/service_instances
//   Action Name: Get
//   Top Level Endpoint: service_instances
//   Return Value: []ServiceInstances
//
//   Method Name: GetSpaceServiceInstances
//   Endpoint: /v2/spaces/:guid/service_instances
//   Action Name: Get
//   Top Level Endpoint: spaces
//   Return Value: []ServiceInstances
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
package ccv2

import "code.cloudfoundry.org/cli/api/cloudcontroller"

// Warnings are a collection of warnings that the Cloud Controller can return
// back from an API request.
type Warnings []string

// CloudControllerClient is a client that can be used to talk to a Cloud
// Controller's V2 Endpoints.
type CloudControllerClient struct {
	authorizationEndpoint     string
	cloudControllerAPIVersion string
	cloudControllerURL        string
	dopplerEndpoint           string
	loggregatorEndpoint       string
	routingEndpoint           string
	tokenEndpoint             string

	connection cloudcontroller.Connection
}

// NewCloudControllerClient returns a new CloudControllerClient.
func NewCloudControllerClient() *CloudControllerClient {
	return new(CloudControllerClient)
}
