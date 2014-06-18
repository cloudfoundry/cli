package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/configuration"
)

type ApplicationSecurityGroupRepo struct {
	gateway net.Gateway
	configRepo configuration.Reader
}

func NewApplicationSecurityGroupRepo(configRepo configuration.Reader, gateway net.Gateway) ApplicationSecurityGroupRepo {
	return ApplicationSecurityGroupRepo{
		configRepo: configRepo,
		gateway: gateway,
	}
}

func (repo ApplicationSecurityGroupRepo) Create(name string) error {
	path := fmt.Sprintf("%s/v2/app_security_groups", repo.configRepo.ApiEndpoint())
	group := ApplicationSecurityGroupFields{Name: name}
	return repo.gateway.CreateResourceFromStruct(path, group)
}

type ApplicationSecurityGroupFields struct {
	Name string `json:"name"`
}
