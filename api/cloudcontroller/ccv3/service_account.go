package ccv3

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/resources"
)

func (client *Client) CreateServiceAccount(account resources.ServiceAccount) (resources.ServiceAccount, Warnings, error) {
	return resources.ServiceAccount{}, nil, errors.New("service account creation is not implemented")
}
