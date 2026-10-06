package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

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
