package installation

import (
	"path/filepath"

	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type TargetProvider interface {
	NewTarget() (Target, error)
}

type targetProvider struct {
	deploymentStateService biconfig.DeploymentStateService
	uuidGenerator          boshuuid.Generator
	installationsRootPath  string
}

func NewTargetProvider(
	deploymentStateService biconfig.DeploymentStateService,
	uuidGenerator boshuuid.Generator,
	installationsRootPath string,
) TargetProvider {
	return &targetProvider{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
		installationsRootPath:  installationsRootPath,
	}
}

func (p *targetProvider) NewTarget() (Target, error) {
	deploymentState, err := p.deploymentStateService.Load()
	if err != nil {
		return Target{}, bosherr.WrapError(err, "Loading deployment state")
	}

	installationID := deploymentState.InstallationID
	if installationID == "" {
		installationID, err = p.uuidGenerator.Generate()
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Generating installation ID")
		}

		deploymentState.InstallationID = installationID
		err := p.deploymentStateService.Save(deploymentState)
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Saving deployment state")
		}
	}

	return NewTarget(filepath.Join(p.installationsRootPath, installationID)), nil
}
