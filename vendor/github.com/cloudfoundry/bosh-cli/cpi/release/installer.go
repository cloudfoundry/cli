package release

import (
	biinstall "github.com/cloudfoundry/bosh-cli/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	birel "github.com/cloudfoundry/bosh-cli/release"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CpiInstaller struct {
	ReleaseManager   birel.Manager
	InstallerFactory biinstall.InstallerFactory
	Validator        Validator
}

func (i CpiInstaller) ValidateCpiRelease(installationManifest biinstallmanifest.Manifest, stage biui.Stage) error {
	return stage.Perform("Validating cpi release", func() error {
		cpiReleaseName := installationManifest.Template.Release
		cpiRelease, found := i.ReleaseManager.Find(cpiReleaseName)
		if !found {
			return bosherr.Errorf("installation release '%s' must refer to a provided release", cpiReleaseName)
		}

		err := i.Validator.Validate(cpiRelease, installationManifest.Template.Name)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", cpiReleaseName)
		}
		return nil
	})
}

func (i CpiInstaller) installCpiRelease(installer biinstall.Installer, installationManifest biinstallmanifest.Manifest, target biinstall.Target, stage biui.Stage) (biinstall.Installation, error) {
	var installation biinstall.Installation
	var err error
	err = stage.PerformComplex("installing CPI", func(installStage biui.Stage) error {
		installation, err = installer.Install(installationManifest, installStage)
		return err
	})
	if err != nil {
		return installation, bosherr.WrapError(err, "Installing CPI")
	}

	return installation, nil
}

func (i CpiInstaller) WithInstalledCpiRelease(installationManifest biinstallmanifest.Manifest, target biinstall.Target, stage biui.Stage, fn func(biinstall.Installation) error) (errToReturn error) {
	installer := i.InstallerFactory.NewInstaller(target)

	installation, err := i.installCpiRelease(installer, installationManifest, target, stage)
	if err != nil {
		errToReturn = err
		return
	}

	defer func() {
		err = i.cleanupInstall(installation, installer, stage)
		if errToReturn == nil {
			errToReturn = err
		}
	}()

	errToReturn = fn(installation)
	return
}

func (i CpiInstaller) cleanupInstall(installation biinstall.Installation, installer biinstall.Installer, stage biui.Stage) error {
	return stage.Perform("Cleaning up rendered CPI jobs", func() error {
		return installer.Cleanup(installation)
	})
}
