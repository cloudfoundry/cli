package command

import (
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	AccessToken() string
	AddPlugin(configv3.Plugin)
	AddPluginRepository(name string, url string)
	APIVersion() string
	BinaryName() string
	BinaryVersion() string
	ColorEnabled() configv3.ColorSetting
	CurrentUser() (configv3.User, error)
	DialTimeout() time.Duration
	DockerPassword() string
	Experimental() bool
	GetPlugin(pluginName string) (configv3.Plugin, bool)
	GetPluginCaseInsensitive(pluginName string) (configv3.Plugin, bool)
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	Locale() string
	MinCLIVersion() string
	NOAARequestRetryCount() int
	OverallPollingTimeout() time.Duration
	PluginHome() string
	PluginRepositories() []configv3.PluginRepository
	Plugins() []configv3.Plugin
	PollingInterval() time.Duration
	RefreshToken() string
	RemovePlugin(string)
	RequestRetryCount() int
	SetAccessToken(token string)
	SetOrganizationInformation(guid string, name string)
	SetRefreshToken(token string)
	SetSpaceInformation(guid string, name string, allowSSH bool)
	SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SetUAAEndpoint(uaaEndpoint string)
	SkipSSLValidation() bool
	SSHOAuthClient() string
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	Target() string
	TargetedOrganization() configv3.Organization
	TargetedSpace() configv3.Space
	UAAOAuthClient() string
	UAAOAuthClientSecret() string
	UnsetOrganizationInformation()
	UnsetSpaceInformation()
	Verbose() (bool, []string)
	WritePluginConfig() error
}
