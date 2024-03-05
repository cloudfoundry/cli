package helpers

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
)

// SkipIfPrivateDockerInfoNotSet skips the test if CF_INT_DOCKER_IMAGE,
// CF_INT_DOCKER_USERNAME, or CF_INT_DOCKER_PASSWORD environment variables are
// not defined.
func SkipIfPrivateDockerInfoNotSet() (string, string, string) {
	privateDockerImage := os.Getenv("CF_INT_DOCKER_IMAGE")
	privateDockerUsername := os.Getenv("CF_INT_DOCKER_USERNAME")
	privateDockerPassword := os.Getenv("CF_INT_DOCKER_PASSWORD")

	if privateDockerImage == "" || privateDockerUsername == "" || privateDockerPassword == "" {
		Skip("CF_INT_DOCKER_IMAGE, CF_INT_DOCKER_USERNAME, or CF_INT_DOCKER_PASSWORD is not set")
	}

	return privateDockerImage, privateDockerUsername, privateDockerPassword
}
