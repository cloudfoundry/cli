package helpers

import (
	"os"

	. "github.com/onsi/ginkgo"
)

func SkipIfClientCredentialsNotSet() (string, string) {
	privateClientID := os.Getenv("CF_INT_CLIENT_ID")
	privateClientSecret := os.Getenv("CF_INT_CLIENT_SECRET")

	if privateClientID == "" || privateClientSecret == "" {
		Skip("CF_INT_CLIENT_ID or CF_INT_CLIENT_SECRET is not set")
	}

	return privateClientID, privateClientSecret
}

func SkipIfCustomClientCredentialsNotSet() (string, string) {
	customClientID := os.Getenv("CF_INT_CUSTOM_CLIENT_ID")
	customClientSecret := os.Getenv("CF_INT_CUSTOM_CLIENT_SECRET")

	if customClientID == "" || customClientSecret == "" {
		Skip("CF_INT_CUSTOM_CLIENT_ID or CF_INT_CUSTOM_CLIENT_SECRET is not set")
	}

	return customClientID, customClientSecret
}
