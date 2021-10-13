// Package cfnetv1 represents a CF Networking V1 client.
//
// These sets of packages are still under development/pre-pre-pre...alpha. Use
// at your own risk! Functionality and design may change without warning.
//
// For more information on the CF Networking API see
// https://github.com/cloudfoundry-incubator/cf-networking-release/blob/develop/docs/API.md
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
//   Return Value: []ServiceInstance
//
//   Method Name: GetSpaceServiceInstances
//   Endpoint: /v2/spaces/:guid/service_instances
//   Action Name: Get
//   Top Level Endpoint: spaces
//   Return Value: []ServiceInstance
//
// Use the following table to determine which HTTP Command equates to which
// Action Name:
//   HTTP Command -> Action Name
//   POST -> Create
//   GET -> Get
//   PUT -> Update
//   DELETE -> Delete
//
// Method Locations
//
// Methods exist in the same file as their return type, regardless of which
// endpoint they use.
//
// Error Handling
//
// All error handling that requires parsing the error_code/code returned back
// from the Cloud Controller should be placed in the errorWrapper. Everything
// else can be handled in the individual operations. All parsed cloud
// controller errors should exist in errors.go, all generic HTTP errors should
// exist in the cloudcontroller's errors.go. Errors related to the individaul
// operation should exist at the top of that operation's file.
//
// No inline-relations-depth And summary Endpoints
//
// This package will not use ever use 'inline-relations-depth' or the
// '/summary' endpoints for any operations. These requests can be extremely
// taxing on the Cloud Controller and are avoided at all costs. Additionally,
// the objects returned back from these requests can become extremely
// inconsistant across versions and are problematic to deal with in general.
package cfnetv1

import (
	"fmt"
	"runtime"
	"time"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1/internal"

	"github.com/tedsuo/rata"
)

// Client is a client that can be used to talk to a CF Networking API.
type Client struct {
	connection cfnetworking.Connection
	router     *rata.RequestGenerator
	url        string
	userAgent  string
}

// Config allows the Client to be configured
type Config struct {
	// AppName is the name of the application/process using the client.
	AppName string

	// AppVersion is the version of the application/process using the client.
	AppVersion string

	// DialTimeout is the DNS timeout used to make all requests to the Cloud
	// Controller.
	DialTimeout time.Duration

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation bool

	// URL is a fully qualified URL to the CF Networking API.
	URL string

	// Wrappers that apply to the client connection.
	Wrappers []ConnectionWrapper
}

// NewClient returns a new CF Networking client.
func NewClient(config Config) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)", config.AppName, config.AppVersion, runtime.Version(), runtime.GOARCH, runtime.GOOS)

	connection := cfnetworking.NewConnection(cfnetworking.Config{
		DialTimeout:       config.DialTimeout,
		SkipSSLValidation: config.SkipSSLValidation,
	})

	wrappedConnection := cfnetworking.NewErrorWrapper().Wrap(connection)
	for _, wrapper := range config.Wrappers {
		wrappedConnection = wrapper.Wrap(wrappedConnection)
	}

	client := &Client{
		connection: wrappedConnection,
		router:     rata.NewRequestGenerator(config.URL, internal.Routes),
		url:        config.URL,
		userAgent:  userAgent,
	}

	return client
}
