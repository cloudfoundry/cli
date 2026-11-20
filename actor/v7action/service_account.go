package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor Actor) GetServiceAccountByNameAndSpace(name, spaceGUID string) (resources.ServiceAccount, Warnings, error) {
	accounts, warnings, err := actor.CloudControllerClient.GetServiceAccounts(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{name}},
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	)
	allWarnings := Warnings(warnings)
	if err != nil {
		return resources.ServiceAccount{}, allWarnings, err
	}
	if len(accounts) == 0 {
		return resources.ServiceAccount{}, allWarnings, fmt.Errorf("Service account '%s' not found in the targeted space.", name)
	}
	if len(accounts) != 1 {
		return resources.ServiceAccount{}, allWarnings, fmt.Errorf("Service account '%s' resolved to multiple accounts.", name)
	}
	account, warnings, err := actor.CloudControllerClient.GetServiceAccount(accounts[0].GUID)
	return account, append(allWarnings, warnings...), err
}

func (actor Actor) DeleteServiceAccountByNameAndSpace(name, spaceGUID string) (Warnings, error) {
	account, warnings, err := actor.GetServiceAccountByNameAndSpace(name, spaceGUID)
	if err != nil {
		return warnings, err
	}
	job, mutationWarnings, err := actor.CloudControllerClient.DeleteServiceAccount(account.GUID)
	return actor.completeServiceAccountOperation(job, append(warnings, mutationWarnings...), err)
}

func (actor Actor) SetServiceAccountEnabledByNameAndSpace(name, spaceGUID string, enabled bool) (Warnings, error) {
	account, warnings, err := actor.GetServiceAccountByNameAndSpace(name, spaceGUID)
	if err != nil {
		return warnings, err
	}
	job, mutationWarnings, err := actor.CloudControllerClient.UpdateServiceAccountEnabled(account.GUID, enabled)
	return actor.completeServiceAccountOperation(job, append(warnings, mutationWarnings...), err)
}

func (actor Actor) BindServiceAccountByNameAndSpace(appName, accountName, spaceGUID string) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	account, accountWarnings, err := actor.GetServiceAccountByNameAndSpace(accountName, spaceGUID)
	warnings = append(warnings, accountWarnings...)
	if err != nil {
		return warnings, err
	}
	job, mutationWarnings, err := actor.CloudControllerClient.UpdateApplicationServiceAccount(app.GUID, resources.Relationship{GUID: account.GUID})
	return actor.completeServiceAccountOperation(job, append(warnings, mutationWarnings...), err)
}

func (actor Actor) UnbindServiceAccountByAppNameAndSpace(appName, spaceGUID string) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}
	job, mutationWarnings, err := actor.CloudControllerClient.UpdateApplicationServiceAccount(app.GUID, resources.Relationship{})
	return actor.completeServiceAccountOperation(job, append(warnings, mutationWarnings...), err)
}

func (actor Actor) completeServiceAccountOperation(job ccv3.JobURL, warnings Warnings, err error) (Warnings, error) {
	if err != nil || job == "" {
		return warnings, err
	}
	pollWarnings, err := actor.CloudControllerClient.PollJob(job)
	return append(warnings, pollWarnings...), err
}

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
