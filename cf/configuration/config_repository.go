package configuration

import (
	"github.com/cloudfoundry/cli/cf/models"
	"sync"
)

type configRepository struct {
	data      *Data
	mutex     *sync.RWMutex
	initOnce  *sync.Once
	persistor Persistor
	onError   func(error)
}

func NewRepositoryFromFilepath(filepath string, errorHandler func(error)) Repository {
	return NewRepositoryFromPersistor(NewDiskPersistor(filepath), errorHandler)
}

func NewRepositoryFromPersistor(persistor Persistor, errorHandler func(error)) Repository {
	c := new(configRepository)
	c.mutex = new(sync.RWMutex)
	c.initOnce = new(sync.Once)
	c.persistor = persistor
	c.onError = errorHandler
	return c
}

type Reader interface {
	ApiEndpoint() string
	ApiVersion() string
	HasAPIEndpoint() bool

	AuthenticationEndpoint() string
	LoggregatorEndpoint() string
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
}

type ReadWriter interface {
	Reader
	ClearSession()
	SetApiEndpoint(string)
	SetApiVersion(string)
	SetAuthenticationEndpoint(string)
	SetLoggregatorEndpoint(string)
	SetUaaEndpoint(string)
	SetAccessToken(string)
	SetRefreshToken(string)
	SetOrganizationFields(models.OrganizationFields)
	SetSpaceFields(models.SpaceFields)
	SetSSLDisabled(bool)
}

type Repository interface {
	ReadWriter
	Close()
}

// ACCESS CONTROL

func (c *configRepository) init() {
	c.initOnce.Do(func() {
		var err error
		c.data, err = c.persistor.Load()
		if err != nil {
			c.onError(err)
		}
	})
}

func (c *configRepository) read(cb func()) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.init()

	cb()
}

func (c *configRepository) write(cb func()) {
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

func (c *configRepository) Close() {
	c.read(func() {
		// perform a read to ensure write lock has been cleared
	})
}

// GETTERS

func (c *configRepository) ApiVersion() (apiVersion string) {
	c.read(func() {
		apiVersion = c.data.ApiVersion
	})
	return
}

func (c *configRepository) AuthenticationEndpoint() (authEndpoint string) {
	c.read(func() {
		authEndpoint = c.data.AuthorizationEndpoint
	})
	return
}

func (c *configRepository) LoggregatorEndpoint() (logEndpoint string) {
	c.read(func() {
		logEndpoint = c.data.LoggregatorEndPoint
	})
	return
}

func (c *configRepository) UaaEndpoint() (uaaEndpoint string) {
	c.read(func() {
		uaaEndpoint = c.data.UaaEndpoint
	})
	return
}

func (c *configRepository) ApiEndpoint() (apiEndpoint string) {
	c.read(func() {
		apiEndpoint = c.data.Target
	})
	return
}

func (c *configRepository) HasAPIEndpoint() (hasEndpoint bool) {
	c.read(func() {
		hasEndpoint = c.data.ApiVersion != "" && c.data.Target != ""
	})
	return
}

func (c *configRepository) AccessToken() (accessToken string) {
	c.read(func() {
		accessToken = c.data.AccessToken
	})
	return
}

func (c *configRepository) RefreshToken() (refreshToken string) {
	c.read(func() {
		refreshToken = c.data.RefreshToken
	})
	return
}

func (c *configRepository) OrganizationFields() (org models.OrganizationFields) {
	c.read(func() {
		org = c.data.OrganizationFields
	})
	return
}

func (c *configRepository) SpaceFields() (space models.SpaceFields) {
	c.read(func() {
		space = c.data.SpaceFields
	})
	return
}

func (c *configRepository) UserEmail() (email string) {
	c.read(func() {
		email = NewTokenInfo(c.data.AccessToken).Email
	})
	return
}

func (c *configRepository) UserGuid() (guid string) {
	c.read(func() {
		guid = NewTokenInfo(c.data.AccessToken).UserGuid
	})
	return
}

func (c *configRepository) Username() (name string) {
	c.read(func() {
		name = NewTokenInfo(c.data.AccessToken).Username
	})
	return
}

func (c *configRepository) IsLoggedIn() (loggedIn bool) {
	c.read(func() {
		loggedIn = c.data.AccessToken != ""
	})
	return
}

func (c *configRepository) HasOrganization() (hasOrg bool) {
	c.read(func() {
		hasOrg = c.data.OrganizationFields.Guid != "" && c.data.OrganizationFields.Name != ""
	})
	return
}

func (c *configRepository) HasSpace() (hasSpace bool) {
	c.read(func() {
		hasSpace = c.data.SpaceFields.Guid != "" && c.data.SpaceFields.Name != ""
	})
	return
}

func (c *configRepository) IsSSLDisabled() (isSSLDisabled bool) {
	c.read(func() {
		isSSLDisabled = c.data.SSLDisabled
	})
	return
}

// SETTERS

func (c *configRepository) ClearSession() {
	c.write(func() {
		c.data.AccessToken = ""
		c.data.RefreshToken = ""
		c.data.OrganizationFields = models.OrganizationFields{}
		c.data.SpaceFields = models.SpaceFields{}
	})
}

func (c *configRepository) SetApiEndpoint(endpoint string) {
	c.write(func() {
		c.data.Target = endpoint
	})
}

func (c *configRepository) SetApiVersion(version string) {
	c.write(func() {
		c.data.ApiVersion = version
	})
}

func (c *configRepository) SetAuthenticationEndpoint(endpoint string) {
	c.write(func() {
		c.data.AuthorizationEndpoint = endpoint
	})
}

func (c *configRepository) SetLoggregatorEndpoint(endpoint string) {
	c.write(func() {
		c.data.LoggregatorEndPoint = endpoint
	})
}

func (c *configRepository) SetUaaEndpoint(uaaEndpoint string) {
	c.write(func() {
		c.data.UaaEndpoint = uaaEndpoint
	})
}

func (c *configRepository) SetAccessToken(token string) {
	c.write(func() {
		c.data.AccessToken = token
	})
}

func (c *configRepository) SetRefreshToken(token string) {
	c.write(func() {
		c.data.RefreshToken = token
	})
}

func (c *configRepository) SetOrganizationFields(org models.OrganizationFields) {
	c.write(func() {
		c.data.OrganizationFields = org
	})
}

func (c *configRepository) SetSpaceFields(space models.SpaceFields) {
	c.write(func() {
		c.data.SpaceFields = space
	})
}

func (c *configRepository) SetSSLDisabled(disabled bool) {
	c.write(func() {
		c.data.SSLDisabled = disabled
	})
}
