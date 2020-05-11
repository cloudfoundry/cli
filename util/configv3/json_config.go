package configv3

import (
	"time"

	"github.com/SermoDigital/jose/jws"
)

// JSONConfig represents .cf/config.json.
type JSONConfig struct {
	AccessToken              string             `json:"AccessToken"`
	APIVersion               string             `json:"APIVersion"`
	AsyncTimeout             int                `json:"AsyncTimeout"`
	AuthorizationEndpoint    string             `json:"AuthorizationEndpoint"`
	ColorEnabled             string             `json:"ColorEnabled"`
	ConfigVersion            int                `json:"ConfigVersion"`
	DopplerEndpoint          string             `json:"DopplerEndPoint"`
	Locale                   string             `json:"Locale"`
	LogCacheEndpoint         string             `json:"LogCacheEndPoint"`
	MinCLIVersion            string             `json:"MinCLIVersion"`
	MinRecommendedCLIVersion string             `json:"MinRecommendedCLIVersion"`
	TargetedOrganization     Organization       `json:"OrganizationFields"`
	PluginRepositories       []PluginRepository `json:"PluginRepos"`
	RefreshToken             string             `json:"RefreshToken"`
	RoutingEndpoint          string             `json:"RoutingAPIEndpoint"`
	TargetedSpace            Space              `json:"SpaceFields"`
	SSHOAuthClient           string             `json:"SSHOAuthClient"`
	SkipSSLValidation        bool               `json:"SSLDisabled"`
	Target                   string             `json:"Target"`
	Trace                    string             `json:"Trace"`
	UAAEndpoint              string             `json:"UaaEndpoint"`
	UAAGrantType             string             `json:"UAAGrantType"`
	UAAOAuthClient           string             `json:"UAAOAuthClient"`
	UAAOAuthClientSecret     string             `json:"UAAOAuthClientSecret"`
}

// Organization contains basic information about the targeted organization.
type Organization struct {
	GUID string `json:"GUID"`
	Name string `json:"Name"`
}

// Space contains basic information about the targeted space.
type Space struct {
	GUID     string `json:"GUID"`
	Name     string `json:"Name"`
	AllowSSH bool   `json:"AllowSSH"`
}

// User represents the user information provided by the JWT access token.
type User struct {
	Name     string
	GUID     string
	Origin   string
	IsClient bool
}

// AccessToken returns the access token for making authenticated API calls.
func (config *Config) AccessToken() string {
	return config.ConfigFile.AccessToken
}

// APIVersion returns the CC API Version.
func (config *Config) APIVersion() string {
	return config.ConfigFile.APIVersion
}

// CurrentUser returns user information decoded from the JWT access token in
// .cf/config.json.
func (config *Config) CurrentUser() (User, error) {
	return decodeUserFromJWT(config.ConfigFile.AccessToken)
}

// CurrentUserName returns the name of a user as returned by CurrentUser()
func (config *Config) CurrentUserName() (string, error) {
	user, err := config.CurrentUser()
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

// HasTargetedOrganization returns true if the organization is set.
func (config *Config) HasTargetedOrganization() bool {
	return config.ConfigFile.TargetedOrganization.GUID != ""
}

// HasTargetedSpace returns true if the space is set.
func (config *Config) HasTargetedSpace() bool {
	return config.ConfigFile.TargetedSpace.GUID != ""
}

// MinCLIVersion returns the minimum CLI version required by the CC.
func (config *Config) MinCLIVersion() string {
	return config.ConfigFile.MinCLIVersion
}

// OverallPollingTimeout returns the overall polling timeout for async
// operations. The time is based off of:
//   1. The config file's AsyncTimeout value (integer) is > 0
//   2. Defaults to the DefaultOverallPollingTimeout
func (config *Config) OverallPollingTimeout() time.Duration {
	if config.ConfigFile.AsyncTimeout == 0 {
		return DefaultOverallPollingTimeout
	}
	return time.Duration(config.ConfigFile.AsyncTimeout) * time.Minute
}

// RefreshToken returns the refresh token for getting a new access token.
func (config *Config) RefreshToken() string {
	return config.ConfigFile.RefreshToken
}

// RoutingEndpoint returns the endpoint for the router API
func (config *Config) RoutingEndpoint() string {
	return config.ConfigFile.RoutingEndpoint
}

// SetAsyncTimeout sets the async timeout.
func (config *Config) SetAsyncTimeout(timeout int) {
	config.ConfigFile.AsyncTimeout = timeout
}

// SetAccessToken sets the current access token.
func (config *Config) SetAccessToken(accessToken string) {
	config.ConfigFile.AccessToken = accessToken
}

// SetColorEnabled sets the color enabled feature to true or false
func (config *Config) SetColorEnabled(enabled string) {
	config.ConfigFile.ColorEnabled = enabled
}

// SetLocale sets the locale, or clears the field if requested
func (config *Config) SetLocale(locale string) {
	if locale == "CLEAR" {
		config.ConfigFile.Locale = ""
	} else {
		config.ConfigFile.Locale = locale
	}
}

// SetMinCLIVersion sets the minimum CLI version required by the CC.
func (config *Config) SetMinCLIVersion(minVersion string) {
	config.ConfigFile.MinCLIVersion = minVersion
}

// SetOrganizationInformation sets the currently targeted organization.
func (config *Config) SetOrganizationInformation(guid string, name string) {
	config.ConfigFile.TargetedOrganization.GUID = guid
	config.ConfigFile.TargetedOrganization.Name = name
}

// SetRefreshToken sets the current refresh token.
func (config *Config) SetRefreshToken(refreshToken string) {
	config.ConfigFile.RefreshToken = refreshToken
}

// SetSpaceInformation sets the currently targeted space.
// The "AllowSSH" field is not returned by v3, and is never read from the config.
// Persist `true` to maintain compatibility in the config file.
// TODO: this field should be removed entirely in v7
func (config *Config) SetSpaceInformation(guid string, name string, allowSSH bool) {
	config.V7SetSpaceInformation(guid, name)
	config.ConfigFile.TargetedSpace.AllowSSH = allowSSH
}

type TargetInformationArgs struct {
	Api               string
	ApiVersion        string
	Auth              string
	Doppler           string
	LogCache          string
	MinCLIVersion     string
	Routing           string
	SkipSSLValidation bool
}

// SetTargetInformation sets the currently targeted CC API and related other
// related API URLs.
func (config *Config) SetTargetInformation(args TargetInformationArgs) {
	config.ConfigFile.Target = args.Api
	config.ConfigFile.APIVersion = args.ApiVersion
	config.ConfigFile.AuthorizationEndpoint = args.Auth
	config.SetMinCLIVersion(args.MinCLIVersion)
	config.ConfigFile.DopplerEndpoint = args.Doppler
	config.ConfigFile.LogCacheEndpoint = args.LogCache
	config.ConfigFile.RoutingEndpoint = args.Routing
	config.ConfigFile.SkipSSLValidation = args.SkipSSLValidation

	config.UnsetOrganizationAndSpaceInformation()
}

// SetTokenInformation sets the current token/user information.
func (config *Config) SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string) {
	config.ConfigFile.AccessToken = accessToken
	config.ConfigFile.RefreshToken = refreshToken
	config.ConfigFile.SSHOAuthClient = sshOAuthClient
}

// SetTrace sets the trace field to either true, false, or a path to a file.
func (config *Config) SetTrace(trace string) {
	config.ConfigFile.Trace = trace
}

// SetUAAClientCredentials sets the client credentials.
func (config *Config) SetUAAClientCredentials(client string, clientSecret string) {
	config.ConfigFile.UAAOAuthClient = client
	config.ConfigFile.UAAOAuthClientSecret = clientSecret
}

// SetUAAEndpoint sets the UAA endpoint that is obtained from hitting
// <AuthorizationEndpoint>/login.
func (config *Config) SetUAAEndpoint(uaaEndpoint string) {
	config.ConfigFile.UAAEndpoint = uaaEndpoint
}

// SetUAAGrantType sets the UAA grant type for logging in and refreshing the
// token.
func (config *Config) SetUAAGrantType(uaaGrantType string) {
	config.ConfigFile.UAAGrantType = uaaGrantType
}

// SkipSSLValidation returns whether or not to skip SSL validation when
// targeting an API endpoint.
func (config *Config) SkipSSLValidation() bool {
	return config.ConfigFile.SkipSSLValidation
}

// SSHOAuthClient returns the OAuth client id used for SSHing into
// application/process containers.
func (config *Config) SSHOAuthClient() string {
	return config.ConfigFile.SSHOAuthClient
}

// Target returns the CC API URL.
func (config *Config) Target() string {
	return config.ConfigFile.Target
}

// TargetedOrganization returns the currently targeted organization.
func (config *Config) TargetedOrganization() Organization {
	return config.ConfigFile.TargetedOrganization
}

// TargetedOrganizationName returns the name of the targeted organization.
func (config *Config) TargetedOrganizationName() string {
	return config.TargetedOrganization().Name
}

// TargetedSpace returns the currently targeted space.
func (config *Config) TargetedSpace() Space {
	return config.ConfigFile.TargetedSpace
}

// UAAGrantType returns the grant type of the supplied UAA credentials.
func (config *Config) UAAGrantType() string {
	return config.ConfigFile.UAAGrantType
}

// UAAOAuthClient returns the CLI's UAA client ID.
func (config *Config) UAAOAuthClient() string {
	return config.ConfigFile.UAAOAuthClient
}

// UAAOAuthClientSecret returns the CLI's UAA client secret.
func (config *Config) UAAOAuthClientSecret() string {
	return config.ConfigFile.UAAOAuthClientSecret
}

// UnsetOrganizationAndSpaceInformation resets the organization and space
// values to default.
func (config *Config) UnsetOrganizationAndSpaceInformation() {
	config.SetOrganizationInformation("", "")
	config.UnsetSpaceInformation()
}

// UnsetSpaceInformation resets the space values to default.
func (config *Config) UnsetSpaceInformation() {
	config.SetSpaceInformation("", "", false)
}

// UnsetUserInformation resets the access token, refresh token, UAA grant type,
// UAA client credentials, and targeted org/space information.
func (config *Config) UnsetUserInformation() {
	config.SetAccessToken("")
	config.SetRefreshToken("")
	config.SetUAAGrantType("")
	config.SetUAAClientCredentials(DefaultUAAOAuthClient, DefaultUAAOAuthClientSecret)

	config.UnsetOrganizationAndSpaceInformation()

}

// V7SetSpaceInformation sets the currently targeted space.
func (config *Config) V7SetSpaceInformation(guid string, name string) {
	config.ConfigFile.TargetedSpace.GUID = guid
	config.ConfigFile.TargetedSpace.Name = name
}

func decodeUserFromJWT(accessToken string) (User, error) {
	if accessToken == "" {
		return User{}, nil
	}

	token, err := jws.ParseJWT([]byte(accessToken[7:]))
	if err != nil {
		return User{}, err
	}

	claims := token.Claims()

	var name, GUID, origin string
	var isClient bool
	if claims.Has("user_name") {
		name = claims.Get("user_name").(string)
		GUID = claims.Get("user_id").(string)
		origin = claims.Get("origin").(string)
		isClient = false
	} else {
		name = claims.Get("client_id").(string)
		GUID = name
		isClient = true
	}

	return User{
		Name:     name,
		GUID:     GUID,
		Origin:   origin,
		IsClient: isClient,
	}, nil
}
