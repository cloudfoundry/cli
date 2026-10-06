package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (client *Client) GetServiceAccount(guid string) (resources.ServiceAccount, Warnings, error) {
	var account resources.ServiceAccount
	_, warnings, err := client.MakeRequest(RequestParams{RequestName: internal.GetServiceAccountRequest, URIParams: internal.Params{"service_account_guid": guid}, ResponseBody: &account})
	return account, warnings, err
}

func (client *Client) DeleteServiceAccount(guid string) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{RequestName: internal.DeleteServiceAccountRequest, URIParams: internal.Params{"service_account_guid": guid}})
}

func (client *Client) UpdateServiceAccountEnabled(guid string, enabled bool) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{RequestName: internal.PatchServiceAccountRequest, URIParams: internal.Params{"service_account_guid": guid}, RequestBody: struct {
		Enabled bool `json:"enabled"`
	}{enabled}})
}

func (client *Client) UpdateApplicationServiceAccount(appGUID string, relationship resources.Relationship) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{RequestName: internal.PatchApplicationServiceAccountRequest, URIParams: internal.Params{"app_guid": appGUID}, RequestBody: relationship})
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
