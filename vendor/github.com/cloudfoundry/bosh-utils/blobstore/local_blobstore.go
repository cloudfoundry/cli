package blobstore

import (
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const (
	blobstorePathPermissions = os.FileMode(0770)
)

type localBlobstore struct {
	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator
	options map[string]interface{}
}

func NewLocalBlobstore(
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	options map[string]interface{},
) Blobstore {
	return localBlobstore{
		fs:      fs,
		uuidGen: uuidGen,
		options: options,
	}
}

func (b localBlobstore) Get(blobID string) (fileName string, err error) {
	file, err := b.fs.TempFile("bosh-blobstore-external-Get")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary file")
	}
	defer file.Close()

	fileName = file.Name()

	err = b.fs.CopyFile(path.Join(b.path(), blobID), fileName)
	if err != nil {
		b.fs.RemoveAll(fileName)
		return "", bosherr.WrapError(err, "Copying file")
	}

	return fileName, nil
}

func (b localBlobstore) CleanUp(fileName string) error {
	b.fs.RemoveAll(fileName)
	return nil
}

func (b localBlobstore) Delete(blobID string) error {
	blobPath := path.Join(b.path(), blobID)
	return b.fs.RemoveAll(blobPath)
}

func (b localBlobstore) Create(fileName string) (blobID string, err error) {
	blobID, err = b.uuidGen.Generate()
	if err != nil {
		err = bosherr.WrapError(err, "Generating blobID")
		return
	}

	err = b.fs.MkdirAll(b.path(), blobstorePathPermissions)
	if err != nil {
		err = bosherr.WrapError(err, "Making blobstore path")
		blobID = ""
		return
	}

	err = b.fs.CopyFile(fileName, path.Join(b.path(), blobID))
	if err != nil {
		err = bosherr.WrapError(err, "Copying file to blobstore path")
		blobID = ""
		return
	}
	return
}

func (b localBlobstore) Validate() error {
	path, found := b.options["blobstore_path"]
	if !found {
		return bosherr.Error("missing blobstore_path")
	}

	_, ok := path.(string)
	if !ok {
		return bosherr.Error("blobstore_path must be a string")
	}

	return nil
}

func (b localBlobstore) path() string {
	// Validate() makes sure that it's a string
	return b.options["blobstore_path"].(string)
}
