package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadingWithNoConfigFile(t *testing.T) {
	config := loadDefaultConfig(t)

	assert.Equal(t, config.Target, "https://api.run.pivotal.io")
	assert.Equal(t, config.ApiVersion, "2")
	assert.Equal(t, config.AuthorizationEndpoint, "https://login.run.pivotal.io")
	assert.Equal(t, config.AccessToken, "")
}

func TestSavingAndLoading(t *testing.T) {
	configToSave := loadDefaultConfig(t)

	configToSave.ApiVersion = "3.1.0"
	configToSave.Target = "https://api.target.example.com"
	configToSave.AuthorizationEndpoint = "https://login.target.example.com"
	configToSave.AccessToken = "bearer my_access_token"

	Save()

	singleton = nil
	savedConfig := Get()
	assert.Equal(t, savedConfig, configToSave)
}

func TestUserEmailWithAValidAccessToken(t *testing.T) {
	config := loadDefaultConfig(t)

	config.AccessToken = "bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E"
	assert.Equal(t, config.UserEmail(), "user1@example.com")
}

func TestUserEmailWithInvalidAccessToken(t *testing.T) {
	config := loadDefaultConfig(t)

	config.AccessToken = "bearer"
	assert.Empty(t, config.UserEmail())

	config.AccessToken = "bearer eyJhbGciOiJSUzI1NiJ9"
	assert.Empty(t, config.UserEmail())
}

func loadDefaultConfig(t *testing.T) (config *Configuration) {
	Delete()
	config = Get()
	return
}
