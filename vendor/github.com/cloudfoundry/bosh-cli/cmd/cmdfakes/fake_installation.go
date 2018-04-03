package cmdfakes

import (
	biinstallation "github.com/cloudfoundry/bosh-cli/installation"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type FakeInstallation struct {
}

func (f *FakeInstallation) Target() biinstallation.Target {
	return biinstallation.Target{}
}

func (f *FakeInstallation) Job() biinstallation.InstalledJob {
	return biinstallation.InstalledJob{}
}

func (f *FakeInstallation) WithRunningRegistry(logger boshlog.Logger, stage biui.Stage, fn func() error) error {
	return fn()
}

func (f *FakeInstallation) StartRegistry() error {
	return nil
}

func (f *FakeInstallation) StopRegistry() error {
	return nil
}
