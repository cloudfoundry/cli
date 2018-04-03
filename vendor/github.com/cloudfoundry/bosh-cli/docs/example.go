package main

import (
	"fmt"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

// Fill out consts and run this file: `go run docs/example.go`

const (
	uaaURL      = "" // eg "https://some-uaa:8443"
	directorURL = "" // eg "https://some-director"

	uaaClient       = "" // eg "my-script"
	uaaClientSecret = "" // eg "my-script-secret"

	someCA = "" /* eg `
	-----BEGIN CERTIFICATE-----
	MIIDXzCCAkegAwIBAgIJAJLKKzS3Z2x3MA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
	...
	vBCS0L9jrvon5LfRJi4xsyAwut5xX98kC3adNJw9RqZApGVKeYfoP5DqcR5vf6vY
	-----END CERTIFICATE-----
	` */
)

func main() {
	uaa, err := buildUAA()
	if err != nil {
		panic(err)
	}

	director, err := buildDirector(uaa)
	if err != nil {
		panic(err)
	}

	// Fetch information about the Director.
	info, err := director.Info()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Director: %s\n", info.Name)

	// See director/interfaces.go for a full list of methods.
	deps, err := director.Deployments()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nDeployments:\n")

	for _, dep := range deps {
		fmt.Printf("- %s\n", dep.Name())
	}
}

func buildUAA() (boshuaa.UAA, error) {
	logger := boshlog.NewLogger(boshlog.LevelError)
	factory := boshuaa.NewFactory(logger)

	// Build a UAA config from a URL.
	// HTTPS is required and certificates are always verified.
	config, err := boshuaa.NewConfigFromURL(uaaURL)
	if err != nil {
		return nil, err
	}

	// Set client credentials for authentication.
	// Machine level access should typically use a client instead of a particular user.
	config.Client = uaaClient
	config.ClientSecret = uaaClientSecret

	// Configure trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = someCA

	return factory.New(config)
}

func buildDirector(uaa boshuaa.UAA) (boshdir.Director, error) {
	logger := boshlog.NewLogger(boshlog.LevelError)
	factory := boshdir.NewFactory(logger)

	// Build a Director config from address-like string.
	// HTTPS is required and certificates are always verified.
	config, err := boshdir.NewConfigFromURL(directorURL)
	if err != nil {
		return nil, err
	}

	// Configure custom trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = someCA

	// Allow Director to fetch UAA tokens when necessary.
	config.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc

	return factory.New(config, boshdir.NewNoopTaskReporter(), boshdir.NewNoopFileReporter())
}
