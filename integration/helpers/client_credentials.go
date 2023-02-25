package helpers

import (
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	if ClientCredentialsTestMode() {
		Skip("CF_INT_CLIENT_CREDENTIALS_TEST_MODE is enabled")
	}
}

func ClientCredentialsTestMode() bool {
	envVar := os.Getenv("CF_INT_CLIENT_CREDENTIALS_TEST_MODE")

	if envVar == "" {
		return false
	}

	testMode, err := strconv.ParseBool(envVar)
	Expect(err).ToNot(HaveOccurred(), "CF_INT_CLIENT_CREDENTIALS_TEST_MODE should be boolean")

	return testMode
}
