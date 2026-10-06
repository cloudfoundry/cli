package v7action

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) CreateServiceAccountInSpace(name, description, spaceGUID string) (resources.ServiceAccount, Warnings, error) {
	return resources.ServiceAccount{}, nil, errors.New("service account creation is not implemented")
}
