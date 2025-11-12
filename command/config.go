package command

import (
	"time"

	"code.cloudfoundry.org/cli/v9/util/configv3"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Config

// Config a way of getting basic CF configuration
type Config interface {
	AccessToken() string
	AddPlugin(configv3.Plugin)
	AddPluginRepository(name string, url string)
	AuthorizationEndpoint() string
	APIVersion() string
	B3TraceID() string
	BinaryName() string
	BinaryVersion() string
	CFPassword() string
	CFUsername() string
	ColorEnabled() configv3.ColorSetting
	CurrentUser() (configv3.User, error)
	CurrentUserName() (string, error)
	DialTimeout() time.Duration
	DockerPassword() string
	CNBCredentials() (map[string]interface{}, error)
	Experimental() bool
	GetPlugin(pluginName string) (configv3.Plugin, bool)
	GetPluginCaseInsensitive(pluginName string) (configv3.Plugin, bool)
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	IsTTY() bool
	Locale() string
	LogCacheEndpoint() string
	MinCLIVersion() string
	NOAARequestRetryCount() int
	NetworkPolicyV1Endpoint() string
	OverallPollingTimeout() time.Duration
	PluginHome() string
	PluginRepositories() []configv3.PluginRepository
	Plugins() []configv3.Plugin
	PollingInterval() time.Duration
	RefreshToken() string
	RemovePlugin(string)
	RequestRetryCount() int
	RoutingEndpoint() string
	SetAsyncTimeout(timeout int)
	SetAccessToken(token string)
	SetColorEnabled(enabled string)
	SetLocale(locale string)
	SetMinCLIVersion(version string)
	SetOrganizationInformation(guid string, name string)
	SetRefreshToken(token string)
	SetSpaceInformation(guid string, name string, allowSSH bool)
	V7SetSpaceInformation(guid string, name string)
	SetTargetInformation(args configv3.TargetInformationArgs)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SetTrace(trace string)
	SetUAAClientCredentials(client string, clientSecret string)
	SetUAAEndpoint(uaaEndpoint string)
	SetUAAGrantType(uaaGrantType string)
	SkipSSLValidation() bool
	SSHOAuthClient() string
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	// TODO: Rename to APITarget()
	Target() string
	TargetedOrganization() configv3.Organization
	TargetedOrganizationName() string
	TargetedSpace() configv3.Space
	TerminalWidth() int
	UAADisableKeepAlives() bool
	UAAEndpoint() string
	UAAGrantType() string
	UAAOAuthClient() string
	UAAOAuthClientSecret() string
	UnsetOrganizationAndSpaceInformation()
	UnsetSpaceInformation()
	UnsetUserInformation()
	Verbose() (bool, []string)
	WritePluginConfig() error
	WriteConfig() error
	IsCFOnK8s() bool
	SetKubernetesAuthInfo(authInfo string)
}
