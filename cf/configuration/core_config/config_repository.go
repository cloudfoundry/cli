package core_config

import (
	"strings"
	"sync"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
)

type ConfigRepository struct {
	data      *Data
	mutex     *sync.RWMutex
	initOnce  *sync.Once
	persistor configuration.Persistor
	onError   func(error)
}

func NewRepositoryFromFilepath(filepath string, errorHandler func(error)) Repository {
	return NewRepositoryFromPersistor(configuration.NewDiskPersistor(filepath), errorHandler)
}

func NewRepositoryFromPersistor(persistor configuration.Persistor, errorHandler func(error)) Repository {
	data := NewData()
	if !persistor.Exists() {
		//set default plugin repo
		data.PluginRepos = append(data.PluginRepos, models.PluginRepo{
			Name: "CF-Community",
			Url:  "http://plugins.cloudfoundry.org",
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
	ApiEndpoint() string
	ApiVersion() string
	HasAPIEndpoint() bool

	AuthenticationEndpoint() string
	LoggregatorEndpoint() string
	DopplerEndpoint() string
	UaaEndpoint() string
	AccessToken() string
	RefreshToken() string

	OrganizationFields() models.OrganizationFields
	HasOrganization() bool

	SpaceFields() models.SpaceFields
	HasSpace() bool

	Username() string
	UserGuid() string
	UserEmail() string
	IsLoggedIn() bool
	IsSSLDisabled() bool
	IsMinApiVersion(string) bool
	IsMinCliVersion(string) bool
	MinCliVersion() string
	MinRecommendedCliVersion() string

	AsyncTimeout() uint
	Trace() string

	ColorEnabled() string

	Locale() string

	PluginRepos() []models.PluginRepo
}

type ReadWriter interface {
	Reader
	ClearSession()
	SetApiEndpoint(string)
	SetApiVersion(string)
	SetMinCliVersion(string)
	SetMinRecommendedCliVersion(string)
	SetAuthenticationEndpoint(string)
	SetLoggregatorEndpoint(string)
	SetDopplerEndpoint(string)
	SetUaaEndpoint(string)
	SetAccessToken(string)
	SetRefreshToken(string)
	SetOrganizationFields(models.OrganizationFields)
	SetSpaceFields(models.SpaceFields)
	SetSSLDisabled(bool)
	SetAsyncTimeout(uint)
	SetTrace(string)
	SetColorEnabled(string)
	SetLocale(string)
	SetPluginRepo(models.PluginRepo)
	UnSetPluginRepo(int)
}

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

func (c *ConfigRepository) ApiVersion() (apiVersion string) {
	c.read(func() {
		apiVersion = c.data.ApiVersion
	})
	return
}

func (c *ConfigRepository) AuthenticationEndpoint() (authEndpoint string) {
	c.read(func() {
		authEndpoint = c.data.AuthorizationEndpoint
	})
	return
}

func (c *ConfigRepository) LoggregatorEndpoint() (logEndpoint string) {
	c.read(func() {
		logEndpoint = c.data.LoggregatorEndPoint
	})
	return
}

func (c *ConfigRepository) DopplerEndpoint() (logEndpoint string) {
	//revert this in v7.0, once CC advertise doppler endpoint, and
	//everyone has migrated from loggregator to doppler

	// c.read(func() {
	// 	logEndpoint = c.data.DopplerEndPoint
	// })
	c.read(func() {
		logEndpoint = c.data.LoggregatorEndPoint
	})

	return strings.Replace(logEndpoint, "loggregator", "doppler", 1)
}

func (c *ConfigRepository) UaaEndpoint() (uaaEndpoint string) {
	c.read(func() {
		uaaEndpoint = c.data.UaaEndpoint
	})
	return
}

func (c *ConfigRepository) ApiEndpoint() (apiEndpoint string) {
	c.read(func() {
		apiEndpoint = c.data.Target
	})
	return
}

func (c *ConfigRepository) HasAPIEndpoint() (hasEndpoint bool) {
	c.read(func() {
		hasEndpoint = c.data.ApiVersion != "" && c.data.Target != ""
	})
	return
}

func (c *ConfigRepository) AccessToken() (accessToken string) {
	c.read(func() {
		accessToken = c.data.AccessToken
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

func (c *ConfigRepository) UserGuid() (guid string) {
	c.read(func() {
		guid = NewTokenInfo(c.data.AccessToken).UserGuid
	})
	return
}

func (c *ConfigRepository) Username() (name string) {
	c.read(func() {
		name = NewTokenInfo(c.data.AccessToken).Username
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
		hasOrg = c.data.OrganizationFields.Guid != "" && c.data.OrganizationFields.Name != ""
	})
	return
}

func (c *ConfigRepository) HasSpace() (hasSpace bool) {
	c.read(func() {
		hasSpace = c.data.SpaceFields.Guid != "" && c.data.SpaceFields.Name != ""
	})
	return
}

func (c *ConfigRepository) IsSSLDisabled() (isSSLDisabled bool) {
	c.read(func() {
		isSSLDisabled = c.data.SSLDisabled
	})
	return
}

func (c *ConfigRepository) IsMinApiVersion(v string) bool {
	var apiVersion string
	c.read(func() {
		apiVersion = c.data.ApiVersion
	})
	return apiVersion >= v
}

func (c *ConfigRepository) IsMinCliVersion(version string) bool {
	if version == "BUILT_FROM_SOURCE" {
		return true
	}
	var minCliVersion string
	c.read(func() {
		minCliVersion = c.data.MinCliVersion
	})
	return version >= minCliVersion
}

func (c *ConfigRepository) MinCliVersion() (minCliVersion string) {
	c.read(func() {
		minCliVersion = c.data.MinCliVersion
	})
	return
}

func (c *ConfigRepository) MinRecommendedCliVersion() (minRecommendedCliVersion string) {
	c.read(func() {
		minRecommendedCliVersion = c.data.MinRecommendedCliVersion
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

func (c *ConfigRepository) SetApiEndpoint(endpoint string) {
	c.write(func() {
		c.data.Target = endpoint
	})
}

func (c *ConfigRepository) SetApiVersion(version string) {
	c.write(func() {
		c.data.ApiVersion = version
	})
}

func (c *ConfigRepository) SetMinCliVersion(version string) {
	c.write(func() {
		c.data.MinCliVersion = version
	})
}

func (c *ConfigRepository) SetMinRecommendedCliVersion(version string) {
	c.write(func() {
		c.data.MinRecommendedCliVersion = version
	})
}

func (c *ConfigRepository) SetAuthenticationEndpoint(endpoint string) {
	c.write(func() {
		c.data.AuthorizationEndpoint = endpoint
	})
}

func (c *ConfigRepository) SetLoggregatorEndpoint(endpoint string) {
	c.write(func() {
		c.data.LoggregatorEndPoint = endpoint
	})
}

func (c *ConfigRepository) SetDopplerEndpoint(endpoint string) {
	c.write(func() {
		c.data.DopplerEndPoint = endpoint
	})
}

func (c *ConfigRepository) SetUaaEndpoint(uaaEndpoint string) {
	c.write(func() {
		c.data.UaaEndpoint = uaaEndpoint
	})
}

func (c *ConfigRepository) SetAccessToken(token string) {
	c.write(func() {
		c.data.AccessToken = token
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
