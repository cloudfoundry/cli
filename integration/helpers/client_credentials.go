package helpers

import (
	"os"

	. "github.com/onsi/ginkgo"
)

// SkipIfClientCredentialsNotSet will skip the test when either
// CF_INT_CLIENT_ID or CF_INT_CLIENT_SECRET are not set.
func SkipIfClientCredentialsNotSet() (string, string) {
	privateClientID := os.Getenv("CF_INT_CLIENT_ID")
	privateClientSecret := os.Getenv("CF_INT_CLIENT_SECRET")

	if privateClientID == "" || privateClientSecret == "" {
		Skip("CF_INT_CLIENT_ID or CF_INT_CLIENT_SECRET is not set")
	}

	return privateClientID, privateClientSecret
}

// SkipIfCustomClientCredentialsNotSet will skip the test when either
// CF_INT_CUSTOM_CLIENT_ID or CF_INT_CUSTOM_CLIENT_SECRET are not set.
func SkipIfCustomClientCredentialsNotSet() (string, string) {
	customClientID := os.Getenv("CF_INT_CUSTOM_CLIENT_ID")
	customClientSecret := os.Getenv("CF_INT_CUSTOM_CLIENT_SECRET")

	if customClientID == "" || customClientSecret == "" {
		Skip("CF_INT_CUSTOM_CLIENT_ID or CF_INT_CUSTOM_CLIENT_SECRET is not set")
	}

	return customClientID, customClientSecret
}

func SkipIfClientCredentialsTestMode() {
	clientCredentialsTestMode := os.Getenv("CF_INT_CLIENT_CREDENTIALS_TEST_MODE")

	if clientCredentialsTestMode != "" {
		Skip("CF_INT_CLIENT_CREDENTIALS_TEST_MODE is set")
	}
}
