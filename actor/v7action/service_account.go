package v7action

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) GetServiceAccountsInSpace(spaceGUID string) ([]resources.ServiceAccount, Warnings, error) {
	accounts, warnings, err := actor.CloudControllerClient.GetServiceAccounts(ccv3.Query{
		Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID},
	})
	return accounts, Warnings(warnings), err
}

func (actor Actor) CreateServiceAccountInSpace(name, description, spaceGUID string) (resources.ServiceAccount, Warnings, error) {
	account, warnings, err := actor.CloudControllerClient.CreateServiceAccount(resources.ServiceAccount{
		Name:          name,
		Description:   description,
		Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: spaceGUID}},
	})
	return account, Warnings(warnings), err
}
