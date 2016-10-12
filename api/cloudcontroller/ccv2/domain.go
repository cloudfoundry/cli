package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Domain struct {
	GUID string
	Name string
}

func (domain *Domain) UnmarshalJSON(data []byte) error {
	var ccDomain struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccDomain); err != nil {
		return err
	}

	domain.GUID = ccDomain.Metadata.GUID
	domain.Name = ccDomain.Entity.Name
	return nil
}

func (client *CloudControllerClient) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.SharedDomainRequest,
		Params:      map[string]string{"shared_domain_guid": domainGUID},
	}

	var domain Domain
	response := cloudcontroller.Response{
		Result: &domain,
	}

	err := client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	return domain, response.Warnings, nil
}

func (client *CloudControllerClient) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.PrivateDomainRequest,
		Params:      map[string]string{"private_domain_guid": domainGUID},
	}

	var domain Domain
	response := cloudcontroller.Response{
		Result: &domain,
	}

	err := client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	return domain, response.Warnings, nil
}
