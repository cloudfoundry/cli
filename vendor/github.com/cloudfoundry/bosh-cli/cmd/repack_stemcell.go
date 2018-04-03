package cmd

import (
	"github.com/cloudfoundry/bosh-cli/stemcell"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type RepackStemcellCmd struct {
	ui                boshui.UI
	fs                boshsys.FileSystem
	stemcellExtractor stemcell.Extractor
}

func NewRepackStemcellCmd(
	ui boshui.UI,
	fs boshsys.FileSystem,
	stemcellExtractor stemcell.Extractor,
) RepackStemcellCmd {
	return RepackStemcellCmd{ui: ui, fs: fs, stemcellExtractor: stemcellExtractor}
}

func (c RepackStemcellCmd) Run(opts RepackStemcellOpts) error {
	extractedStemcell, err := c.stemcellExtractor.Extract(opts.Args.PathToStemcell)
	if err != nil {
		return err
	}

	if opts.Name != "" {
		extractedStemcell.SetName(opts.Name)
	}

	if opts.Version != "" {
		extractedStemcell.SetVersion(opts.Version)
	}

	if opts.EmptyImage {
		err = extractedStemcell.EmptyImage()
		if err != nil {
			return err
		}
	}

	if opts.CloudProperties != "" {
		cloudProperties := new(biproperty.Map)
		err = yaml.Unmarshal([]byte(opts.CloudProperties), cloudProperties)
		if err != nil {
			return err
		}

		extractedStemcell.SetCloudProperties(*cloudProperties)
	}

	if len(opts.Format) != 0 {
		extractedStemcell.SetFormat(opts.Format)
	}

	return extractedStemcell.Pack(opts.Args.PathToResult.ExpandedPath)
}
