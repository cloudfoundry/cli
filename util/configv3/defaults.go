package configv3

import "time"

const (
	// DefaultDialTimeout is the default timeout for the dail.
	DefaultDialTimeout = 5 * time.Second

	// DefaultNOAARetryCount is the default number of request retries.
	DefaultNOAARetryCount = 5

	// DefaultOverallPollingTimeout is the default maximum time that the CLI will
	// poll a job running on the Cloud Controller. By default it's infinit, which
	// is represented by MaxInt64.
	DefaultOverallPollingTimeout = time.Duration(1 << 62)
	// Developer Note: Due to bugs in using MaxInt64 during comparison, the above
	// was chosen as a replacement.

	// DefaultPollingInterval is the time between consecutive polls of a status.
	DefaultPollingInterval = 3 * time.Second

	// DefaultStagingTimeout is the default timeout for application staging.
	DefaultStagingTimeout = 15 * time.Minute

	// DefaultStartupTimeout is the default timeout for application starting.
	DefaultStartupTimeout = 5 * time.Minute

	// DefaultTarget is the default CFConfig value for Target.
	DefaultTarget = ""

	// DefaultSSHOAuthClient is the default oauth client ID for SSHing into an
	// application/process container
	DefaultSSHOAuthClient = "ssh-proxy"

	// DefaultUAAOAuthClient is the default client ID for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClient = "cf"

	// DefaultUAAOAuthClientSecret is the default client secret for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClientSecret = ""

	// DefaultRetryCount is the default number of request retries.
	DefaultRetryCount = 2
)

// NOAARequestRetryCount returns the number of request retries.
func (*Config) NOAARequestRetryCount() int {
	return DefaultNOAARetryCount
}

// PollingInterval returns the time between polls.
func (config *Config) PollingInterval() time.Duration {
	return DefaultPollingInterval
}

// RequestRetryCount returns the number of request retries.
func (*Config) RequestRetryCount() int {
	return DefaultRetryCount
}
