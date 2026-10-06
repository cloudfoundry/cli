package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

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
