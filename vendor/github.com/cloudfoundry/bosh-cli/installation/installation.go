package installation

import (
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	biregistry "github.com/cloudfoundry/bosh-cli/registry"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Installation interface {
	Target() Target
	Job() InstalledJob
	WithRunningRegistry(boshlog.Logger, biui.Stage, func() error) error
	StartRegistry() error
	StopRegistry() error
}

type installation struct {
	target                Target
	job                   InstalledJob
	manifest              biinstallmanifest.Manifest
	registryServerManager biregistry.ServerManager

	registryServer biregistry.Server
}

func NewInstallation(
	target Target,
	job InstalledJob,
	manifest biinstallmanifest.Manifest,
	registryServerManager biregistry.ServerManager,
) Installation {
	return &installation{
		target:                target,
		job:                   job,
		manifest:              manifest,
		registryServerManager: registryServerManager,
	}
}

func (i *installation) Target() Target {
	return i.target
}

func (i *installation) Job() InstalledJob {
	return i.job
}

func (i *installation) WithRunningRegistry(logger boshlog.Logger, stage biui.Stage, fn func() error) error {
	err := stage.Perform("Starting registry", func() error {
		return i.StartRegistry()
	})
	if err != nil {
		return err
	}
	defer i.stopRegistryNice(logger, stage)
	return fn()
}

func (i *installation) stopRegistryNice(logger boshlog.Logger, stage biui.Stage) {
	err := stage.Perform("Stopping registry", func() error {
		return i.StopRegistry()
	})
	if err != nil {
		logger.Warn("installation", "Registry failed to stop: %s", err)
	}
}

func (i *installation) StartRegistry() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer != nil {
			return bosherr.Error("Registry already started")
		}
		config := i.manifest.Registry
		registryServer, err := i.registryServerManager.Start(config.Username, config.Password, config.Host, config.Port)
		if err != nil {
			return bosherr.WrapError(err, "Starting registry")
		}
		i.registryServer = registryServer
	}
	return nil
}

func (i *installation) StopRegistry() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer == nil {
			return bosherr.Error("Registry must be started before it can be stopped")
		}
		err := i.registryServer.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping registry")
		}
		i.registryServer = nil
	}
	return nil
}
