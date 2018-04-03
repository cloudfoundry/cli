package blobstore

import (
	"fmt"
	"net/http"
	"net/url"

	boshdavcli "github.com/cloudfoundry/bosh-davcli/client"
	boshdavcliconf "github.com/cloudfoundry/bosh-davcli/config"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type Factory interface {
	Create(string, *http.Client) (Blobstore, error)
}

type blobstoreFactory struct {
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
}

func NewBlobstoreFactory(uuidGenerator boshuuid.Generator, fs boshsys.FileSystem, logger boshlog.Logger) Factory {
	return blobstoreFactory{
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
	}
}

func (f blobstoreFactory) Create(blobstoreURL string, httpClient *http.Client) (Blobstore, error) {
	blobstoreConfig, err := f.parseBlobstoreURL(blobstoreURL)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating blobstore config")
	}

	davClient := boshdavcli.NewClient(boshdavcliconf.Config{
		Endpoint: fmt.Sprintf("%s/blobs", blobstoreConfig.Endpoint),
		User:     blobstoreConfig.Username,
		Password: blobstoreConfig.Password,
	}, httpClient, f.logger)

	return NewBlobstore(davClient, f.uuidGenerator, f.fs, f.logger), nil
}

func (f blobstoreFactory) parseBlobstoreURL(blobstoreURL string) (Config, error) {
	parsedURL, err := url.Parse(blobstoreURL)
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password string
	userInfo := parsedURL.User
	if userInfo != nil {
		username = userInfo.Username()
		password, _ = userInfo.Password()
	}

	return Config{
		Endpoint: fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
		Username: username,
		Password: password,
	}, nil
}
