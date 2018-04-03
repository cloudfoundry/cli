package cmd

import (
	"github.com/cloudfoundry/bosh-cli/crypto"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	"github.com/cloudfoundry/bosh-cli/release/license"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
)

type RedigestReleaseCmd struct {
	reader                boshrel.Reader
	writer                boshrel.Writer
	digestCalculator      crypto.DigestCalculator
	mv                    boshfu.Mover
	archiveFilePathReader crypto2.ArchiveDigestFilePathReader
	ui                    boshui.UI
}

func NewRedigestReleaseCmd(
	reader boshrel.Reader,
	writer boshrel.Writer,
	digestCalculator crypto.DigestCalculator,
	mv boshfu.Mover,
	archiveFilePathReader crypto2.ArchiveDigestFilePathReader,
	ui boshui.UI,
) RedigestReleaseCmd {
	return RedigestReleaseCmd{
		reader:           reader,
		writer:           writer,
		digestCalculator: digestCalculator,
		mv:               mv,
		archiveFilePathReader: archiveFilePathReader,
		ui: ui,
	}
}

func (cmd RedigestReleaseCmd) Run(args RedigestReleaseArgs) error {
	release, err := cmd.reader.Read(args.Path)
	if err != nil {
		return err
	}

	newJobs := []*boshjob.Job{}
	for _, job := range release.Jobs() {
		newJob, err := job.RehashWithCalculator(cmd.digestCalculator, cmd.archiveFilePathReader)
		if err != nil {
			return err
		}
		newJobs = append(newJobs, newJob)
	}

	newCompiledPackages := []*boshpkg.CompiledPackage{}
	for _, compPkg := range release.CompiledPackages() {
		newCompiledPackage, err := compPkg.RehashWithCalculator(cmd.digestCalculator, cmd.archiveFilePathReader)
		if err != nil {
			return err
		}
		newCompiledPackages = append(newCompiledPackages, newCompiledPackage)
	}

	newPackages := []*boshpkg.Package{}
	for _, pkg := range release.Packages() {
		newPkg, err := pkg.RehashWithCalculator(cmd.digestCalculator, cmd.archiveFilePathReader)
		if err != nil {
			return err
		}
		newPackages = append(newPackages, newPkg)
	}

	var newLicense *license.License
	releaseLicense := release.License()
	if releaseLicense != nil {
		newLicense, err = releaseLicense.RehashWithCalculator(cmd.digestCalculator, cmd.archiveFilePathReader)
		if err != nil {
			return err
		}
	}

	newRelease := release.CopyWith(newJobs, newPackages, newLicense, newCompiledPackages)

	tmpWriterPath, err := cmd.writer.Write(newRelease, nil)
	if err != nil {
		return err
	}

	err = cmd.mv.Move(tmpWriterPath, args.Destination.ExpandedPath)
	if err != nil {
		return err
	}

	ReleaseTables{Release: newRelease, ArchivePath: args.Destination.ExpandedPath}.Print(cmd.ui)

	return nil
}
