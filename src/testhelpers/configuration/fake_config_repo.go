package configuration

import (
	"cf/configuration"
	"cf"
)

var TestConfigurationSingleton *configuration.Configuration
var SavedConfiguration configuration.Configuration

type FakeConfigRepository struct {
}

func (repo FakeConfigRepository) SetOrganization(org cf.OrganizationFields) (err error) {
	config, err := repo.Get()
	if err != nil{
		return
	}

	config.OrganizationFields = org
	config.SpaceFields = cf.SpaceFields{}
	return repo.Save()
}

func (repo FakeConfigRepository) SetSpace(space cf.SpaceFields) (err error) {
	config, err := repo.Get()
	if err != nil {
		return
	}

	config.SpaceFields = space
	return repo.Save()
}

func (repo FakeConfigRepository) Get() (c *configuration.Configuration, err error) {
	if TestConfigurationSingleton == nil {
		TestConfigurationSingleton = new(configuration.Configuration)
		TestConfigurationSingleton.Target = "https://api.run.pivotal.io"
		TestConfigurationSingleton.ApiVersion = "2"
		TestConfigurationSingleton.AuthorizationEndpoint = "https://login.run.pivotal.io"
		TestConfigurationSingleton.ApplicationStartTimeout = 30 // seconds
	}

	return TestConfigurationSingleton, nil
}

func (repo FakeConfigRepository) Delete() {
	SavedConfiguration = configuration.Configuration{}
	TestConfigurationSingleton = nil
}

func (repo FakeConfigRepository) Save() (err error) {
	SavedConfiguration = *TestConfigurationSingleton
	return
}

func (repo FakeConfigRepository) ClearTokens() (err error) {
	c, _ := repo.Get()
	c.AccessToken = ""
	c.RefreshToken = ""

	return nil
}

func (repo FakeConfigRepository) ClearSession() (err error) {
	repo.ClearTokens()

	c, _ := repo.Get()
	c.OrganizationFields = cf.OrganizationFields{}
	c.SpaceFields = cf.SpaceFields{}

	return nil
}

func (repo FakeConfigRepository) Login() (c *configuration.Configuration) {
	c, _ = repo.Get()
	c.AccessToken = `BEARER eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E`
	return
}
