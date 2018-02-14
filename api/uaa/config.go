package uaa

import "time"

//go:generate counterfeiter . Config

// Config allows the Client to be configured
type Config interface {
	// BinaryName is the name of the application/process using the client.
	BinaryName() string

	// BinaryVersion is the version of the application/process using the client.
	BinaryVersion() string

	// DialTimeout is the DNS lookup timeout for the client. If not set, it is
	// infinite.
	DialTimeout() time.Duration

	// UAAOAuthClient is the UAA client ID the client will use.
	UAAOAuthClient() string

	// UAAOAuthClientSecret is the UAA client secret the client will use.
	UAAOAuthClientSecret() string

	// SetUAAEndpoint sets the UAA endpoint that is obtained from hitting
	// <AuthorizationEndpoint>/login.
	SetUAAEndpoint(uaaEndpoint string)

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation() bool
}
