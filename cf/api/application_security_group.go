package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
)

type AppSecurityGroup interface {
	Create(ApplicationSecurityGroupFields) error
}

type ApplicationSecurityGroupRepo struct {
	gateway    net.Gateway
	configRepo configuration.Reader
}

func NewApplicationSecurityGroupRepo(configRepo configuration.Reader, gateway net.Gateway) ApplicationSecurityGroupRepo {
	return ApplicationSecurityGroupRepo{
		configRepo: configRepo,
		gateway:    gateway,
	}
}

func (repo ApplicationSecurityGroupRepo) Create(groupFields ApplicationSecurityGroupFields) error {
	path := fmt.Sprintf("%s/v2/app_security_groups", repo.configRepo.ApiEndpoint())
	return repo.gateway.CreateResourceFromStruct(path, groupFields)
}

type ApplicationSecurityGroupFields struct {
	Name       string              `json:"name"`
	Rules      []map[string]string `json:"rules"`
	SpaceGuids []string            `json:"space_guids"`
}
