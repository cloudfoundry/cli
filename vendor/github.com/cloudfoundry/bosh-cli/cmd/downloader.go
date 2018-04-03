package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/clock"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type Downloader interface {
	Download(blobstoreID, sha1, prefix, dstDirPath string) error
}

type UIDownloader struct {
	director    boshdir.Director
	timeService clock.Clock

	fs boshsys.FileSystem
	ui biui.UI
}

func NewUIDownloader(
	director boshdir.Director,
	timeService clock.Clock,
	fs boshsys.FileSystem,
	ui biui.UI,
) UIDownloader {
	return UIDownloader{
		director:    director,
		timeService: timeService,

		fs: fs,
		ui: ui,
	}
}

func (d UIDownloader) Download(blobstoreID, sha1, prefix, dstDirPath string) error {
	tsSuffix := strings.Replace(d.timeService.Now().Format("20060102-150405.999999999"), ".", "-", -1)

	dstFileName := fmt.Sprintf("%s-%s.tgz", prefix, tsSuffix)

	dstFilePath := filepath.Join(dstDirPath, dstFileName)

	tmpFile, err := d.fs.TempFile(fmt.Sprintf("director-resource-%s", blobstoreID))
	if err != nil {
		return err
	}

	defer d.fs.RemoveAll(tmpFile.Name())

	d.ui.PrintLinef("Downloading resource '%s' to '%s'...", blobstoreID, dstFilePath)

	err = d.director.DownloadResourceUnchecked(blobstoreID, tmpFile)
	if err != nil {
		return err
	}

	// unfortunate. apparently old directors may not send the digest.
	if len(sha1) > 0 {
		err = d.verifyFile(tmpFile, sha1)
		if err != nil {
			return err
		}
	}

	err = boshfu.NewFileMover(d.fs).Move(tmpFile.Name(), dstFilePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Moving to final destination")
	}

	return nil
}

func (d UIDownloader) verifyFile(file boshsys.File, expectedDigest string) error {
	expectedMultipleDigest, err := boshcrypto.ParseMultipleDigest(expectedDigest)
	if err != nil {
		return err
	}

	return expectedMultipleDigest.VerifyFilePath(file.Name(), d.fs)
}
