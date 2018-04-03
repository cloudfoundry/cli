package cmd

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type ExportReleaseCmd struct {
	deployment boshdir.Deployment
	downloader Downloader
}

func NewExportReleaseCmd(deployment boshdir.Deployment, downloader Downloader) ExportReleaseCmd {
	return ExportReleaseCmd{deployment: deployment, downloader: downloader}
}

func (c ExportReleaseCmd) Run(opts ExportReleaseOpts) error {
	rel := opts.Args.ReleaseSlug
	os := opts.Args.OSVersionSlug
	jobs := opts.Jobs

	result, err := c.deployment.ExportRelease(rel, os, jobs)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s-%s-%s-%s", rel.Name(), rel.Version(), os.OS(), os.Version())

	err = c.downloader.Download(
		result.BlobstoreID,
		result.SHA1,
		prefix,
		opts.Directory.Path,
	)
	if err != nil {
		return bosherr.WrapError(err, "Downloading exported release")
	}

	return nil
}
