package v7action

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
	"errors"
)

func (actor Actor) GetServiceAccountsInSpace(spaceGUID string) ([]resources.ServiceAccount, Warnings, error) {
	return nil, nil, errors.New("service account listing is not implemented")
}

func (actor Actor) CreateServiceAccountInSpace(name, description, spaceGUID string) (resources.ServiceAccount, Warnings, error) {
	account, warnings, err := actor.CloudControllerClient.CreateServiceAccount(resources.ServiceAccount{
		Name:          name,
		Description:   description,
		Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: spaceGUID}},
	})
	return account, Warnings(warnings), err
}
