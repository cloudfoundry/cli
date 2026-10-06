package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
	"errors"
)

func (client *Client) GetServiceAccount(guid string) (resources.ServiceAccount, Warnings, error) {
	return resources.ServiceAccount{}, nil, errors.New("service account show is not implemented")
}

func (client *Client) DeleteServiceAccount(guid string) (JobURL, Warnings, error) {
	return "", nil, errors.New("service account deletion is not implemented")
}

func (client *Client) UpdateServiceAccountEnabled(guid string, enabled bool) (JobURL, Warnings, error) {
	return "", nil, errors.New("service account enablement is not implemented")
}

func (client *Client) UpdateApplicationServiceAccount(appGUID string, relationship resources.Relationship) (JobURL, Warnings, error) {
	return "", nil, errors.New("service account assignment is not implemented")
}

func (client *Client) GetServiceAccounts(query ...Query) ([]resources.ServiceAccount, Warnings, error) {
	var accounts []resources.ServiceAccount
	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceAccountsRequest,
		Query:        query,
		ResponseBody: resources.ServiceAccount{},
		AppendToList: func(item interface{}) error {
			accounts = append(accounts, item.(resources.ServiceAccount))
			return nil
		},
	})
	return accounts, warnings, err
}

func (client *Client) CreateServiceAccount(account resources.ServiceAccount) (resources.ServiceAccount, Warnings, error) {
	// Only platform-independent creation inputs may be sent; identity and
	// provisioning state are supplied by CAPI.
	request := struct {
		Name          string                  `json:"name"`
		Description   string                  `json:"description,omitempty"`
		Relationships resources.Relationships `json:"relationships"`
	}{account.Name, account.Description, account.Relationships}
	var response resources.ServiceAccount
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostServiceAccountsRequest,
		RequestBody:  request,
		ResponseBody: &response,
	})
	return response, warnings, err
}
