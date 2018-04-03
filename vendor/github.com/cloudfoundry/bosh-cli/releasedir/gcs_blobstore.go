package releasedir

import (
	gobytes "bytes"
	"context"
	"encoding/json"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	gcsclient "github.com/cloudfoundry/bosh-gcscli/client"
	gcsconfig "github.com/cloudfoundry/bosh-gcscli/config"
)

type GCSBlobstore struct {
	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator
	options map[string]interface{}
}

func NewGCSBlobstore(
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	options map[string]interface{},
) GCSBlobstore {
	return GCSBlobstore{
		fs:      fs,
		uuidGen: uuidGen,
		options: options,
	}
}

func (b GCSBlobstore) Get(blobID string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	file, err := b.fs.TempFile("bosh-gcs-blob")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating destination file")
	}
	defer file.Close()

	if err := client.Get(blobID, file); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (b GCSBlobstore) Create(path string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	blobID, err := b.uuidGen.Generate()
	if err != nil {
		return "", bosherr.WrapError(err, "Generating blobstore ID")
	}

	file, err := b.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", bosherr.WrapError(err, "Opening source file")
	}
	defer file.Close()

	if err := client.Put(file, blobID); err != nil {
		return "", err
	}

	return blobID, nil
}

func (b GCSBlobstore) CleanUp(path string) error {
	return b.fs.RemoveAll(path)
}

func (b GCSBlobstore) Delete(blobID string) error {
	panic("Not implemented")
}

func (b GCSBlobstore) Validate() error {
	_, err := b.client()
	return err
}

func (b GCSBlobstore) client() (*gcsclient.GCSBlobstore, error) {
	bytes, err := json.Marshal(b.options)
	if err != nil {
		return nil, bosherr.WrapError(err, "Marshaling config")
	}

	conf, err := gcsconfig.NewFromReader(gobytes.NewBuffer(bytes))
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading config")
	}

	client, err := gcsclient.New(context.Background(), &conf)
	if err != nil {
		return nil, bosherr.WrapError(err, "Validating config")
	}

	return client, nil
}
