package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
)

type UploadBlobsCmd struct {
	blobsDir boshreldir.BlobsDir
}

func NewUploadBlobsCmd(blobsDir boshreldir.BlobsDir) UploadBlobsCmd {
	return UploadBlobsCmd{blobsDir: blobsDir}
}

func (c UploadBlobsCmd) Run() error {
	err := c.blobsDir.UploadBlobs()
	if err != nil {
		return bosherr.WrapErrorf(err, "Uploading blobs")
	}

	return nil
}
