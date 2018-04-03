package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentRepo interface {
	UpdateCurrent(manifestSHA string) error
	FindCurrent() (manifestSHA string, found bool, err error)
}

type deploymentRepo struct {
	deploymentStateService DeploymentStateService
}

func NewDeploymentRepo(deploymentStateService DeploymentStateService) DeploymentRepo {
	return deploymentRepo{
		deploymentStateService: deploymentStateService,
	}
}

func (r deploymentRepo) FindCurrent() (string, bool, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return "", false, bosherr.WrapError(err, "Loading existing config")
	}

	currentManifestSHA := deploymentState.CurrentManifestSHA
	if currentManifestSHA != "" {
		return currentManifestSHA, true, nil
	}

	return "", false, nil
}

func (r deploymentRepo) UpdateCurrent(manifestSHA string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	deploymentState.CurrentManifestSHA = manifestSHA

	err = r.deploymentStateService.Save(deploymentState)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}
