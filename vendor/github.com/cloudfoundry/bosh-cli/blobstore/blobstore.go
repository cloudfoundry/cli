package blobstore

import (
	"io"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type Blobstore interface {
	Get(blobID string) (LocalBlob, error)
	Add(sourcePath string) (blobID string, err error)
}

type DavCLIClient interface {
	Get(path string) (content io.ReadCloser, err error)
	Put(path string, content io.ReadCloser, contentLength int64) (err error)
}

type Config struct {
	Endpoint string
	Username string
	Password string
}

type blobstore struct {
	davClient     DavCLIClient
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	logTag        string
}

func NewBlobstore(davClient DavCLIClient, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem, logger boshlog.Logger) Blobstore {
	return &blobstore{
		davClient:     davClient,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
		logTag:        "blobstore",
	}
}

func (b *blobstore) Get(blobID string) (LocalBlob, error) {
	file, err := b.fs.TempFile("bosh-init-local-blob")
	destinationPath := file.Name()
	err = file.Close()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Closing new temp file '%s'", destinationPath)
	}

	b.logger.Debug(b.logTag, "Downloading blob %s to %s", blobID, destinationPath)

	readCloser, err := b.davClient.Get(blobID)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting blob %s from blobstore", blobID)
	}
	defer func() {
		if err = readCloser.Close(); err != nil {
			b.logger.Warn(b.logTag, "Couldn't close davClient.Get reader: %s", err.Error())
		}
	}()

	targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening file for blob at %s", destinationPath)
	}

	_, err = io.Copy(targetFile, readCloser)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Saving blob to %s", destinationPath)
	}

	return NewLocalBlob(destinationPath, b.fs, b.logger), nil
}

func (b *blobstore) Add(sourcePath string) (string, error) {
	blobID, err := b.uuidGenerator.Generate()
	if err != nil {
		return "", bosherr.WrapError(err, "Generating Blob ID")
	}

	b.logger.Debug(b.logTag, "Uploading blob %s from %s", blobID, sourcePath)

	file, err := b.fs.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Opening file for reading %s", sourcePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			b.logger.Warn(b.logTag, "Couldn't close source file: %s", err.Error())
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Getting fileInfo from %s", sourcePath)
	}

	err = b.davClient.Put(blobID, file, fileInfo.Size())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Putting file '%s' into blobstore (via DAVClient) as blobID '%s'", sourcePath, blobID)
	}

	return blobID, nil
}
