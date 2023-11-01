package coreconfig

import (
	"strings"
	"sync"

	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver/v4"
)

type ConfigRepository struct {
	CFCLIVersion string
	data         *Data
	mutex        *sync.RWMutex
	initOnce     *sync.Once
	persistor    configuration.Persistor
	onError      func(error)
}

type CCInfo struct {
	APIVersion               string `json:"api_version"`
	AuthorizationEndpoint    string `json:"authorization_endpoint"`
	DopplerEndpoint          string `json:"doppler_logging_endpoint"`
	LogCacheEndpoint         string `json:"log_cache_endpoint"`
	MinCLIVersion            string `json:"min_cli_version"`
	MinRecommendedCLIVersion string `json:"min_recommended_cli_version"`
	SSHOAuthClient           string `json:"app_ssh_oauth_client"`
	RoutingAPIEndpoint       string `json:"routing_endpoint"`
}

func NewRepositoryFromFilepath(filepath string, errorHandler func(error)) Repository {
	if errorHandler == nil {
		return nil
	}
	return NewRepositoryFromPersistor(configuration.NewDiskPersistor(filepath), errorHandler)
}

func NewRepositoryFromPersistor(persistor configuration.Persistor, errorHandler func(error)) Repository {
	data := NewData()
	if !persistor.Exists() {
		//set default plugin repo
		data.PluginRepos = append(data.PluginRepos, models.PluginRepo{
			Name: "CF-Community",
			URL:  "https://plugins.cloudfoundry.org",
		})
	}

	return &ConfigRepository{
		data:      data,
		mutex:     new(sync.RWMutex),
		initOnce:  new(sync.Once),
		persistor: persistor,
		onError:   errorHandler,
	}
}

type Reader interface {
	APIEndpoint() string
	APIVersion() string
	HasAPIEndpoint() bool

	AuthenticationEndpoint() string
	DopplerEndpoint() string
	LogCacheEndpoint() string
	UaaEndpoint() string
	RoutingAPIEndpoint() string
	AccessToken() string
	UAAOAuthClient() string
	UAAOAuthClientSecret() string
	SSHOAuthClient() string
	RefreshToken() string

	OrganizationFields() models.OrganizationFields
	HasOrganization() bool

	SpaceFields() models.SpaceFields
	HasSpace() bool

	Username() string
	UserGUID() string
	UserEmail() string
	IsLoggedIn() bool
	IsSSLDisabled() bool
	IsMinAPIVersion(semver.Version) bool
	IsMinCLIVersion(string) bool
	MinCLIVersion() string
	MinRecommendedCLIVersion() string
	CLIVersion() string

	AsyncTimeout() uint
	Trace() string

	ColorEnabled() string

	Locale() string

	PluginRepos() []models.PluginRepo
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ReadWriter

type ReadWriter interface {
	Reader
	ClearSession()
	SetAccessToken(string)
	SetAPIEndpoint(string)
	SetAPIVersion(string)
	SetAsyncTimeout(uint)
	SetAuthenticationEndpoint(string)
	SetCLIVersion(string)
	SetColorEnabled(string)
	SetDopplerEndpoint(string)
	SetLogCacheEndpoint(string)
	SetLocale(string)
	SetMinCLIVersion(string)
	SetMinRecommendedCLIVersion(string)
	SetOrganizationFields(models.OrganizationFields)
	SetPluginRepo(models.PluginRepo)
	SetRefreshToken(string)
	SetRoutingAPIEndpoint(string)
	SetSpaceFields(models.SpaceFields)
	SetSSHOAuthClient(string)
	SetSSLDisabled(bool)
	SetTrace(string)
	SetUaaEndpoint(string)
	SetUAAGrantType(string)
	SetUAAOAuthClient(string)
	SetUAAOAuthClientSecret(string)
	UAAGrantType() string
	UnSetPluginRepo(int)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Repository

type Repository interface {
	ReadWriter
	Close()
}

// ACCESS CONTROL

func (c *ConfigRepository) init() {
	c.initOnce.Do(func() {
		err := c.persistor.Load(c.data)
		if err != nil {
			c.onError(err)
		}
	})
}

func (c *ConfigRepository) read(cb func()) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.init()

	cb()
}

func (c *ConfigRepository) write(cb func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.init()

	cb()

	err := c.persistor.Save(c.data)
	if err != nil {
		c.onError(err)
	}
}

// CLOSERS

func (c *ConfigRepository) Close() {
	c.read(func() {
		// perform a read to ensure write lock has been cleared
	})
}

// GETTERS

func (c *ConfigRepository) APIVersion() (apiVersion string) {
	c.read(func() {
		apiVersion = c.data.APIVersion
	})
	return
}

func (c *ConfigRepository) AuthenticationEndpoint() (authEndpoint string) {
	c.read(func() {
		authEndpoint = c.data.AuthorizationEndpoint
	})
	return
}

func (c *ConfigRepository) DopplerEndpoint() (dopplerEndpoint string) {
	c.read(func() {
		dopplerEndpoint = c.data.DopplerEndPoint
	})

	return
}

func (c *ConfigRepository) LogCacheEndpoint() (logCacheEndpoint string) {
	c.read(func() {
		logCacheEndpoint = c.data.LogCacheEndPoint
	})

	return
}

func (c *ConfigRepository) UaaEndpoint() (uaaEndpoint string) {
	c.read(func() {
		uaaEndpoint = c.data.UaaEndpoint
	})
	return
}

func (c *ConfigRepository) RoutingAPIEndpoint() (routingAPIEndpoint string) {
	c.read(func() {
		routingAPIEndpoint = c.data.RoutingAPIEndpoint
	})
	return
}

func (c *ConfigRepository) APIEndpoint() string {
	var apiEndpoint string
	c.read(func() {
		apiEndpoint = c.data.Target
	})
	apiEndpoint = strings.TrimRight(apiEndpoint, "/")

	return apiEndpoint
}

func (c *ConfigRepository) HasAPIEndpoint() (hasEndpoint bool) {
	c.read(func() {
		hasEndpoint = c.data.APIVersion != "" && c.data.Target != ""
	})
	return
}

func (c *ConfigRepository) AccessToken() (accessToken string) {
	c.read(func() {
		accessToken = c.data.AccessToken
	})
	return
}

func (c *ConfigRepository) UAAOAuthClient() (clientID string) {
	c.read(func() {
		clientID = c.data.UAAOAuthClient
	})
	return
}

func (c *ConfigRepository) UAAOAuthClientSecret() (clientID string) {
	c.read(func() {
		clientID = c.data.UAAOAuthClientSecret
	})
	return
}

func (c *ConfigRepository) SSHOAuthClient() (clientID string) {
	c.read(func() {
		clientID = c.data.SSHOAuthClient
	})
	return
}

func (c *ConfigRepository) RefreshToken() (refreshToken string) {
	c.read(func() {
		refreshToken = c.data.RefreshToken
	})
	return
}

func (c *ConfigRepository) OrganizationFields() (org models.OrganizationFields) {
	c.read(func() {
		org = c.data.OrganizationFields
	})
	return
}

func (c *ConfigRepository) SpaceFields() (space models.SpaceFields) {
	c.read(func() {
		space = c.data.SpaceFields
	})
	return
}

func (c *ConfigRepository) UserEmail() (email string) {
	c.read(func() {
		email = NewTokenInfo(c.data.AccessToken).Email
	})
	return
}

func (c *ConfigRepository) UserGUID() (guid string) {
	c.read(func() {
		guid = NewTokenInfo(c.data.AccessToken).UserGUID
	})
	return
}

func (c *ConfigRepository) Username() (name string) {
	c.read(func() {
		t := NewTokenInfo(c.data.AccessToken)
		if t.Username != "" {
			name = t.Username
		} else {
			name = t.ClientID
		}
	})
	return
}

func (c *ConfigRepository) IsLoggedIn() (loggedIn bool) {
	c.read(func() {
		loggedIn = c.data.AccessToken != ""
	})
	return
}

func (c *ConfigRepository) HasOrganization() (hasOrg bool) {
	c.read(func() {
		hasOrg = c.data.OrganizationFields.GUID != "" && c.data.OrganizationFields.Name != ""
	})
	return
}

func (c *ConfigRepository) HasSpace() (hasSpace bool) {
	c.read(func() {
		hasSpace = c.data.SpaceFields.GUID != "" && c.data.SpaceFields.Name != ""
	})
	return
}

func (c *ConfigRepository) IsSSLDisabled() (isSSLDisabled bool) {
	c.read(func() {
		isSSLDisabled = c.data.SSLDisabled
	})
	return
}

// SetCLIVersion should only be used in testing
func (c *ConfigRepository) SetCLIVersion(v string) {
	c.CFCLIVersion = v
}

func (c *ConfigRepository) CLIVersion() string {
	if c.CFCLIVersion == "" {
		return version.VersionString()
	} else {
		return c.CFCLIVersion
	}
}

func (c *ConfigRepository) IsMinAPIVersion(requiredVersion semver.Version) bool {
	var apiVersion string
	c.read(func() {
		apiVersion = c.data.APIVersion
	})

	actualVersion, err := semver.Make(apiVersion)
	if err != nil {
		return false
	}
	return actualVersion.GTE(requiredVersion)
}

func (c *ConfigRepository) IsMinCLIVersion(checkVersion string) bool {
	if checkVersion == version.DefaultVersion {
		return true
	}
	var minCLIVersion string
	c.read(func() {
		minCLIVersion = c.data.MinCLIVersion
	})
	if minCLIVersion == "" {
		return true
	}

	actualVersion, err := semver.Make(checkVersion)
	if err != nil {
		return false
	}
	requiredVersion, err := semver.Make(minCLIVersion)
	if err != nil {
		return false
	}
	return actualVersion.GTE(requiredVersion)
}

func (c *ConfigRepository) MinCLIVersion() (minCLIVersion string) {
	c.read(func() {
		minCLIVersion = c.data.MinCLIVersion
	})
	return
}

func (c *ConfigRepository) MinRecommendedCLIVersion() (minRecommendedCLIVersion string) {
	c.read(func() {
		minRecommendedCLIVersion = c.data.MinRecommendedCLIVersion
	})
	return
}

func (c *ConfigRepository) AsyncTimeout() (timeout uint) {
	c.read(func() {
		timeout = c.data.AsyncTimeout
	})
	return
}

func (c *ConfigRepository) Trace() (trace string) {
	c.read(func() {
		trace = c.data.Trace
	})
	return
}

func (c *ConfigRepository) ColorEnabled() (enabled string) {
	c.read(func() {
		enabled = c.data.ColorEnabled
	})
	return
}

func (c *ConfigRepository) Locale() (locale string) {
	c.read(func() {
		locale = c.data.Locale
	})
	return
}

func (c *ConfigRepository) PluginRepos() (repos []models.PluginRepo) {
	c.read(func() {
		repos = c.data.PluginRepos
	})
	return
}

// SETTERS

func (c *ConfigRepository) ClearSession() {
	c.write(func() {
		c.data.AccessToken = ""
		c.data.RefreshToken = ""
		c.data.OrganizationFields = models.OrganizationFields{}
		c.data.SpaceFields = models.SpaceFields{}
	})
}

func (c *ConfigRepository) SetAPIEndpoint(endpoint string) {
	c.write(func() {
		c.data.Target = endpoint
	})
}

func (c *ConfigRepository) SetAPIVersion(version string) {
	c.write(func() {
		c.data.APIVersion = version
	})
}

func (c *ConfigRepository) SetMinCLIVersion(version string) {
	c.write(func() {
		c.data.MinCLIVersion = version
	})
}

func (c *ConfigRepository) SetMinRecommendedCLIVersion(version string) {
	c.write(func() {
		c.data.MinRecommendedCLIVersion = version
	})
}

func (c *ConfigRepository) SetAuthenticationEndpoint(endpoint string) {
	c.write(func() {
		c.data.AuthorizationEndpoint = endpoint
	})
}

func (c *ConfigRepository) SetDopplerEndpoint(endpoint string) {
	c.write(func() {
		c.data.DopplerEndPoint = endpoint
	})
}

func (c *ConfigRepository) SetLogCacheEndpoint(endpoint string) {
	c.write(func() {
		c.data.LogCacheEndPoint = endpoint
	})
}

func (c *ConfigRepository) SetUaaEndpoint(uaaEndpoint string) {
	c.write(func() {
		c.data.UaaEndpoint = uaaEndpoint
	})
}

func (c *ConfigRepository) SetRoutingAPIEndpoint(routingAPIEndpoint string) {
	c.write(func() {
		c.data.RoutingAPIEndpoint = routingAPIEndpoint
	})
}

func (c *ConfigRepository) SetAccessToken(token string) {
	c.write(func() {
		c.data.AccessToken = token
	})
}

func (c *ConfigRepository) SetUAAOAuthClient(clientID string) {
	c.write(func() {
		c.data.UAAOAuthClient = clientID
	})
}

func (c *ConfigRepository) SetUAAOAuthClientSecret(clientID string) {
	c.write(func() {
		c.data.UAAOAuthClientSecret = clientID
	})
}

func (c *ConfigRepository) SetSSHOAuthClient(clientID string) {
	c.write(func() {
		c.data.SSHOAuthClient = clientID
	})
}

func (c *ConfigRepository) SetRefreshToken(token string) {
	c.write(func() {
		c.data.RefreshToken = token
	})
}

func (c *ConfigRepository) SetOrganizationFields(org models.OrganizationFields) {
	c.write(func() {
		c.data.OrganizationFields = org
	})
}

func (c *ConfigRepository) SetSpaceFields(space models.SpaceFields) {
	c.write(func() {
		c.data.SpaceFields = space
	})
}

func (c *ConfigRepository) SetSSLDisabled(disabled bool) {
	c.write(func() {
		c.data.SSLDisabled = disabled
	})
}

func (c *ConfigRepository) SetAsyncTimeout(timeout uint) {
	c.write(func() {
		c.data.AsyncTimeout = timeout
	})
}

func (c *ConfigRepository) SetTrace(value string) {
	c.write(func() {
		c.data.Trace = value
	})
}

func (c *ConfigRepository) SetColorEnabled(enabled string) {
	c.write(func() {
		c.data.ColorEnabled = enabled
	})
}

func (c *ConfigRepository) SetLocale(locale string) {
	c.write(func() {
		c.data.Locale = locale
	})
}

func (c *ConfigRepository) SetPluginRepo(repo models.PluginRepo) {
	c.write(func() {
		c.data.PluginRepos = append(c.data.PluginRepos, repo)
	})
}

func (c *ConfigRepository) UnSetPluginRepo(index int) {
	c.write(func() {
		c.data.PluginRepos = append(c.data.PluginRepos[:index], c.data.PluginRepos[index+1:]...)
	})
}

func (c *ConfigRepository) UAAGrantType() string {
	grantType := ""
	c.read(func() {
		grantType = c.data.UAAGrantType
	})
	return grantType
}

func (c *ConfigRepository) SetUAAGrantType(grantType string) {
	c.write(func() {
		c.data.UAAGrantType = grantType
	})
}
